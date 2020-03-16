package maoxianbot

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	maoxianv1 "github.com/sunny0826/maoxian-operator/pkg/apis/maoxian/v1"
	"github.com/xanzy/go-gitlab"
)

type GitlabBot struct {
	client   *gitlab.Client
	HookUrl  string
	Repo     string
	Username string
}

// checkAllBots check all bots
func checkAllBots(statusList []maoxianv1.RepoStatus) []maoxianv1.RepoStatus {
	var result []maoxianv1.RepoStatus
	log.Info("Start Check All Bots")
	for _, status := range statusList {
		gitlabBot := GitlabBot{
			client:   gitlabClient(adminAccess, gitUrl),
			HookUrl:  hookUrl,
			Repo:     status.Name,
			Username: username,
		}
		item := gitlabBot.checkGitlabBot()
		result = append(result, item)
	}
	return result
}

// addMultGitlabBots add gitlab bots
func addMultGitlabBots(statusList []maoxianv1.RepoStatus, createList []string) []maoxianv1.RepoStatus {
	log.Info("--Start Add Bots--")
	for _, repo := range createList {
		gitlabBot := GitlabBot{
			client:   gitlabClient(adminAccess, gitUrl),
			HookUrl:  hookUrl,
			Repo:     repo,
			Username: username,
		}
		status := gitlabBot.addGitlabBot()
		statusList = append(statusList, status)
	}
	log.Info("--Finish Add Bots--")
	return statusList
}

// delMultGitlabBots delete gitlab bots
func delMultGitlabBots(statusList []maoxianv1.RepoStatus, delList []string) []maoxianv1.RepoStatus {
	log.Info("--Start Remove Bots--")
	var newStatusList []maoxianv1.RepoStatus
	for _, status := range statusList {
		var delflag bool
		for _, del := range delList {
			if status.Name == del {
				delflag = true
				gitlabBot := GitlabBot{
					client:   gitlabClient(adminAccess, gitUrl),
					HookUrl:  hookUrl,
					Repo:     del,
					Username: username,
				}
				err := gitlabBot.removeGitlabBot()
				if err != nil {
					log.Error(err, "Failure to remove webhook")
				}
				break
			}
		}
		if !delflag {
			newStatusList = append(newStatusList, status)
		}
	}
	log.Info("--Finish Remove Bots--")
	return newStatusList
}

// checkGitlabBot
func (git *GitlabBot) checkGitlabBot() maoxianv1.RepoStatus {
	log.Info("check gitlab bot", "uername", git.Username, "repo", git.Repo)
	status := maoxianv1.RepoStatus{
		Name:   git.Repo,
		Status: "Checking",
	}
	userId, err := getId(git.client, git.Username, git.Repo)
	if err != nil {
		status.Status = "Failure"
		status.Error = err.Error()
		log.Error(err, "git project & user failure")
		return status
	}
	isExit, err := checkMember(git.client, git.Repo, userId)
	if err != nil {
		status.Status = "Failure"
		status.Error = err.Error()
		log.Error(err, "check member failure")
		return status
	}
	if !isExit {
		err = addMember(git.client, git.Repo, userId)
		if err != nil {
			status.Status = "Failure"
			status.Error = err.Error()
			log.Error(err, "add member failure")
			return status
		}
	}
	err = checkWebhook(git.client, git.Repo, git.HookUrl)
	if err != nil {
		status.Status = "Failure"
		status.Error = err.Error()
		log.Error(err, "check webhook failure")
		return status
	}
	status.Error = ""
	status.Status = "Success"
	status.Success = true
	return status
}

// addGitlabBot add gitlab bot
func (git *GitlabBot) addGitlabBot() maoxianv1.RepoStatus {
	status := maoxianv1.RepoStatus{
		Name:   git.Repo,
		Status: "Pending",
	}
	userId, err := getId(git.client, git.Username, git.Repo)
	if err != nil {
		status.Status = "Failure"
		status.Error = err.Error()
		log.Error(err, "git project & user failure")
		return status
	}
	err = addMember(git.client, git.Repo, userId)
	if err != nil {
		status.Status = "Failure"
		status.Error = err.Error()
		log.Error(err, "add member failure")
		return status
	}
	err = addWebhook(git.client, git.Repo, git.HookUrl, webhookToken)
	if err != nil {
		status.Status = "Failure"
		status.Error = err.Error()
		log.Error(err, "add webhook failure")
		return status
	}
	status.Error = ""
	status.Status = "Success"
	status.Success = true
	return status
}

// removeGitlabBot delete gitlab bot
func (git *GitlabBot) removeGitlabBot() error {
	userId, err := getId(git.client, git.Username, git.Repo)
	if err != nil {
		return err
	}
	err = removeMember(git.client, git.Repo, userId)
	if err != nil {
		return err
	}
	err = removeWebhook(git.client, git.Repo, git.HookUrl)
	if err != nil {
		return err
	}
	return nil
}

// gitlabClient client of gitlab
func gitlabClient(secret string, baseUrl string) *gitlab.Client {
	git := gitlab.NewClient(nil, secret)
	url := fmt.Sprintf("%s/api/v4", baseUrl)
	err := git.SetBaseURL(url)
	if err != nil {
		log.Error(err, "gitlabClient error")
	}
	return git
}

// getId get userId
func getId(client *gitlab.Client, username string, repo string) (int, error) {
	userOpt := &gitlab.ListUsersOptions{
		ListOptions: gitlab.ListOptions{},
		Username:    gitlab.String(username),
	}
	names, _, err := client.Users.ListUsers(userOpt)
	if err != nil {
		return 0, err
	}
	if len(names) == 0 {
		return 0, botError{context: fmt.Sprintf("can not find user:「%s」,please create lab user.", username)}
	} else if len(names) > 1 {
		return 0, botError{context: fmt.Sprintf("please check username,find more than one result,username:%s", username)}
	}
	userId := names[0].ID
	return userId, nil
}

// checkMember check gitlab member of project
func checkMember(client *gitlab.Client, project string, userId int) (bool, error) {
	log.Info("check member", "project", project, "userId", userId)
	memberOpt := &gitlab.ListProjectMembersOptions{
		Query: gitlab.String(username),
	}
	members, _, err := client.ProjectMembers.ListAllProjectMembers(project, memberOpt)
	if err != nil {
		return false, err
	}
	for _, member := range members {
		if member.ID == userId {
			log.Info("Member already exists", "Name", member.Name, "project", project)
			return true, nil
		}
	}
	log.Info("Member do not exists", "Project", project, "Name", username)
	return false, nil
}

// removeMember remove gitlab member of project
func removeMember(client *gitlab.Client, project string, userId int) error {
	isExit, err := checkMember(client, project, userId)
	if err != nil {
		return err
	}
	if isExit {
		_, err := client.ProjectMembers.DeleteProjectMember(project, userId)
		if err != nil {
			return err
		}
		log.Info("remove member successful!", "project", project, "userID", userId)
	} else {
		log.Info("member already remove", "project", project)
	}
	return nil
}

// checkAndAddMember check and add gitlab member of project
func addMember(client *gitlab.Client, project string, userId int) error {
	isExit, err := checkMember(client, project, userId)
	if err != nil {
		return err
	}
	if !isExit {
		add_opt := &gitlab.AddProjectMemberOptions{
			UserID:      gitlab.Int(userId),
			AccessLevel: gitlab.AccessLevel(30),
		}
		member, _, err := client.ProjectMembers.AddProjectMember(project, add_opt)
		if err != nil {
			return err
		}
		log.Info("add member successful!", "project", project, "userID", member.ID, "userName", member.Name, "level", member.AccessLevel)
	} else {
		log.Info("member already exists", "project", project)
	}
	return nil
}

// checkWebhook
func checkWebhook(client *gitlab.Client, project string, hookUrl string) error {
	hookOpt := &gitlab.ListProjectHooksOptions{}
	hooks, _, err := client.Projects.ListProjectHooks(project, hookOpt)
	if err != nil {
		return err
	}
	var hookId int
	for _, hook := range hooks {
		if hook.URL == hookUrl {
			log.Info("webhook exists", "url", hookUrl)
			hookId = hook.ID
			break
		}
	}
	if hookId != 0 {
		editOpt := &gitlab.EditProjectHookOptions{
			URL:                   gitlab.String(hookUrl),
			Token:                 gitlab.String(webhookToken),
			NoteEvents:            gitlab.Bool(true),
			PushEvents:            gitlab.Bool(false),
			EnableSSLVerification: gitlab.Bool(false),
		}
		_, _, err := client.Projects.EditProjectHook(project, hookId, editOpt)
		if err != nil {
			return err
		}
		log.Info("update webhook info", "project", project, "hookUrl", hookUrl)
	} else {
		addOpt := &gitlab.AddProjectHookOptions{
			URL:                   gitlab.String(hookUrl),
			Token:                 gitlab.String(webhookToken),
			NoteEvents:            gitlab.Bool(true),
			PushEvents:            gitlab.Bool(false),
			EnableSSLVerification: gitlab.Bool(false),
		}
		_, _, err := client.Projects.AddProjectHook(project, addOpt)
		if err != nil {
			return err
		}
		log.Info("add webhook successful", "url", hookUrl, "token", webhookToken)
	}
	return nil
}

// removeWebhook delete gitlab webhook of project
func removeWebhook(client *gitlab.Client, project string, hookUrl string) error {
	hookOpt := &gitlab.ListProjectHooksOptions{}
	hooks, _, err := client.Projects.ListProjectHooks(project, hookOpt)
	if err != nil {
		return err
	}
	var hookId int
	for _, hook := range hooks {
		if hook.URL == hookUrl {
			log.Info("webhook exists", "url", hookUrl, "project", project)
			hookId = hook.ID
			break
		}
	}
	if hookId != 0 {
		_, err := client.Projects.DeleteProjectHook(project, hookId)
		if err != nil {
			return err
		}
		log.Info("remove webhook successful!", "project", project, "hookId", hookId, "hookUrl", hookUrl)
	} else {
		log.Info("member already remove", "project", project, "hookId", hookId, "hookUrl", hookUrl)
	}
	return nil
}

// addWebhook add gitlab webhook of project
func addWebhook(client *gitlab.Client, project string, hookUrl string, hmacToken string) error {
	hookOpt := &gitlab.ListProjectHooksOptions{}
	hooks, _, err := client.Projects.ListProjectHooks(project, hookOpt)
	if err != nil {
		return err
	}
	var hookIsExit bool
	for _, hook := range hooks {
		if hook.URL == hookUrl {
			log.Info("webhook already exists", "url", hookUrl, "project", project)
			if !hook.NoteEvents {
				log.Info("IssuesEvents is closed")
				editOpt := &gitlab.EditProjectHookOptions{
					URL:                   gitlab.String(hookUrl),
					Token:                 gitlab.String(hmacToken),
					NoteEvents:            gitlab.Bool(true),
					PushEvents:            gitlab.Bool(false),
					EnableSSLVerification: gitlab.Bool(false),
				}
				_, _, err := client.Projects.EditProjectHook(hook.ProjectID, hook.ID, editOpt)
				if err != nil {
					return err
				}
				log.Info("Open IssuesEvents", "project", project, "hookUrl", hookUrl)
			} else {
				log.Info("IssuesEvents already open", "project", project, "hookUrl", hookUrl)
			}
			hookIsExit = true
		}
	}
	if !hookIsExit {
		addOpt := &gitlab.AddProjectHookOptions{
			URL:                   gitlab.String(hookUrl),
			Token:                 gitlab.String(hmacToken),
			NoteEvents:            gitlab.Bool(true),
			PushEvents:            gitlab.Bool(false),
			EnableSSLVerification: gitlab.Bool(false),
		}
		_, _, err := client.Projects.AddProjectHook(project, addOpt)
		if err != nil {
			return err
		}
		log.Info("add webhook successful", "url", hookUrl, "project", project, "token", hmacToken)
	}
	return nil
}

// generateHmac
func generateHmac(secretName string) string {
	secret := "maoxian"
	log.Info("start generateHmac", "Secret", secret, " Data", secretName)
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(secretName))
	sha := hex.EncodeToString(h.Sum(nil))
	return sha
}
