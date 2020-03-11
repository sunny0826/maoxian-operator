package maoxianbot

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	maoxianv1 "github.com/sunny0826/maoxian-operator/pkg/apis/maoxian/v1"
	"github.com/xanzy/go-gitlab"
	"strings"
)

type GitlabBot struct {
	client   *gitlab.Client
	HookUrl  string
	Repo     string
	Username string
}

//func updateMultGitlabBots(repoListStatue maoxianv1.MaoxianBotStatus, updateList []string, webhookToken string) maoxianv1.MaoxianBotStatus {
//	for _, repo := range updateList {
//		gitlabBot := GitlabBot{
//			client:   gitlabClient(adminAccess, gitUrl),
//			HookUrl:  hookUrl,
//			Repo:     repo,
//			Username: username,
//		}
//		for i, status := range repoListStatue.RepoStatus {
//			if status.Name == repo {
//				log.Info("update status", "Status.Name", status.Name)
//				status = gitlabBot.addGitlabBot(webhookToken)
//				repoListStatue.RepoStatus[i] = status
//				break
//			}
//		}
//	}
//	return repoListStatue
//}

func addMultGitlabBots(statusList []maoxianv1.RepoStatus, createList []string, webhookToken string) []maoxianv1.RepoStatus {
	for _, repo := range createList {
		gitlabBot := GitlabBot{
			client:   gitlabClient(adminAccess, gitUrl),
			HookUrl:  hookUrl,
			Repo:     repo,
			Username: username,
		}
		status := gitlabBot.addGitlabBot(webhookToken)
		statusList = append(statusList, status)
	}
	return statusList
}

func delMultGitlabBots(statusList []maoxianv1.RepoStatus, delList []string) []maoxianv1.RepoStatus {
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
	return newStatusList
}

func (git *GitlabBot) addGitlabBot(webhookToken string) maoxianv1.RepoStatus {
	status := maoxianv1.RepoStatus{
		Name:   git.Repo,
		Status: "Pending",
	}
	userId, projectId, err := getId(git.client, git.Username, git.Repo)
	if err != nil {
		status.Status = "Failure"
		status.Error = err.Error()
		return status
	}
	err = checkAndAddMember(git.client, projectId, userId)
	if err != nil {
		status.Status = "Failure"
		status.Error = err.Error()
		return status
	}
	err = addWebhook(git.client, projectId, git.HookUrl, webhookToken)
	if err != nil {
		status.Status = "Failure"
		status.Error = err.Error()
		return status
	}
	status.Error = ""
	status.Status = "Success"
	status.Success = true
	return status
}

func (git *GitlabBot) removeGitlabBot() error {
	userId, projectId, err := getId(git.client, git.Username, git.Repo)
	if err != nil {
		return err
	}
	err = checkAndRemoveMember(git.client, projectId, userId)
	if err != nil {
		return err
	}
	err = removeWebhook(git.client, projectId, git.HookUrl)
	if err != nil {
		return err
	}
	return nil
}

func gitlabClient(secret string, baseUrl string) *gitlab.Client {
	git := gitlab.NewClient(nil, secret)
	url := fmt.Sprintf("%s/api/v4", baseUrl)
	err := git.SetBaseURL(url)
	if err != nil {
		log.Error(err, "gitlabClient error")
	}
	return git
}

func getId(client *gitlab.Client, username string, repo string) (int, int, error) {
	userOpt := &gitlab.ListUsersOptions{
		ListOptions: gitlab.ListOptions{},
		Username:    gitlab.String(username),
	}
	names, _, err := client.Users.ListUsers(userOpt)
	if err != nil {
		return 0, 0, err
	}
	if len(names) == 0 {
		return 0, 0, botError{context: fmt.Sprintf("can not find user:「%s」,please create lab user.", username)}
	} else if len(names) > 1 {
		return 0, 0, botError{context: fmt.Sprintf("please check username,find more than one result,username:%s", username)}
	}
	userId := names[0].ID
	var projectId int
	projectOpt := &gitlab.SearchOptions{}
	repoName := strings.Split(repo, "/")[1]
	projects, _, err := client.Search.Projects(repoName, projectOpt)
	if err != nil {
		return 0, 0, err
	}
	for _, project := range projects {
		if project.PathWithNamespace == repo {
			projectId = project.ID
			break
		}
	}
	if projectId == 0 {
		return 0, 0, botError{context: fmt.Sprintf("can not find project:「%s」", repo)}
	}
	return userId, projectId, nil
}

func checkAndRemoveMember(client *gitlab.Client, projectId int, userId int) error {
	memberOpt := &gitlab.ListProjectMembersOptions{}
	members, _, err := client.ProjectMembers.ListAllProjectMembers(projectId, memberOpt)
	if err != nil {
		return err
	}
	var isExit bool
	for _, member := range members {
		if member.ID == userId {
			isExit = true
			break
		}
	}
	if isExit {
		_, err := client.ProjectMembers.DeleteProjectMember(projectId, userId)
		if err != nil {
			return err
		}
		log.Info("remove member successful!", "userID", userId, "userName")
	} else {
		log.Info("member already remove")
	}
	return nil
}

func checkAndAddMember(client *gitlab.Client, projectId int, userId int) error {
	memberOpt := &gitlab.ListProjectMembersOptions{}
	members, _, err := client.ProjectMembers.ListAllProjectMembers(projectId, memberOpt)
	if err != nil {
		return err
	}
	var isExit bool
	for _, member := range members {
		if member.ID == userId {
			isExit = true
			break
		}
	}
	if !isExit {
		add_opt := &gitlab.AddProjectMemberOptions{
			UserID:      gitlab.Int(userId),
			AccessLevel: gitlab.AccessLevel(20),
		}
		member, _, err := client.ProjectMembers.AddProjectMember(projectId, add_opt)
		if err != nil {
			return err
		}
		log.Info("add member successful!", "userID", member.ID, "userName", member.Name, "level", member.AccessLevel)
	} else {
		log.Info("member already exists")
	}
	return nil
}

func removeWebhook(client *gitlab.Client, projectId int, hookUrl string) error {
	hookOpt := &gitlab.ListProjectHooksOptions{}
	hooks, _, err := client.Projects.ListProjectHooks(projectId, hookOpt)
	if err != nil {
		return err
	}
	var hookId int
	for _, hook := range hooks {
		if hook.URL == hookUrl {
			log.V(-1).Info("webhook exists", "url", hookUrl)
			hookId = hook.ID
			break
		}
	}
	if hookId != 0 {
		_, err := client.Projects.DeleteProjectHook(projectId, hookId)
		if err != nil {
			return err
		}
		log.Info("remove webhook successful!", "projectId", projectId, "hookId", hookId, "hookUrl", hookUrl)
	} else {
		log.Info("member already remove")
	}
	return nil
}

func addWebhook(client *gitlab.Client, projectId int, hookUrl string, hmacToken string) error {
	hookOpt := &gitlab.ListProjectHooksOptions{}
	hooks, _, err := client.Projects.ListProjectHooks(projectId, hookOpt)
	if err != nil {
		return err
	}
	var hookIsExit bool
	for _, hook := range hooks {
		if hook.URL == hookUrl {
			log.V(-1).Info("webhook already exists", "url", hookUrl)
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
				log.Info("Open IssuesEvents")
			} else {
				log.Info("IssuesEvents already open")
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
		_, _, err := client.Projects.AddProjectHook(projectId, addOpt)
		if err != nil {
			return err
		}
		log.Info("add webhook successful", "url", hookUrl)
		log.Info("create success", "token", hmacToken)
	}
	return nil
}

func generateHmac(secretName string) string {
	secret := "maoxian"
	log.Info("start generateHmac", "Secret", secret, " Data", secretName)
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(secretName))
	sha := hex.EncodeToString(h.Sum(nil))
	return sha
}
