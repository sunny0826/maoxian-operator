package maoxianbot

import (
	"context"
	maoxianv1 "github.com/sunny0826/maoxian-operator/pkg/apis/maoxian/v1"
	"gopkg.in/fatih/set.v0"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_maoxianbot")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new MaoxianBot Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileMaoxianBot{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("maoxianbot-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource MaoxianBot
	err = c.Watch(&source.Kind{Type: &maoxianv1.MaoxianBot{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner MaoxianBot
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &maoxianv1.MaoxianBot{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileMaoxianBot implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileMaoxianBot{}

// ReconcileMaoxianBot reconciles a MaoxianBot object
type ReconcileMaoxianBot struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

var (
	adminAccess string
	username    string
	gitUrl      string
	hookUrl     string
	secretName  string
)

// Reconcile reads that state of the cluster for a MaoxianBot object and makes changes based on the state read
// and what is in the MaoxianBot.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileMaoxianBot) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling MaoxianBot")
	forget := reconcile.Result{}
	adminAccess = os.Getenv("ADMIN_ACCESS")
	username = os.Getenv("BOT_USER")
	gitUrl = os.Getenv("GIT_URL")
	hookUrl = os.Getenv("WEBHOOK")
	secretName = os.Getenv("SECREC_NAME")

	// Fetch the MaoxianBot instance
	instance := &maoxianv1.MaoxianBot{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return forget, nil
		}
		// Error reading the object - requeue the request.
		return forget, err
	}

	if len(instance.Spec.RepoList) == 0 {
		return forget, nil
	}

	if err := controllerutil.SetControllerReference(instance, instance, r.scheme); err != nil {
		return forget, err
	}

	repoList := instance.Spec.RepoList
	repoListStatue := instance.Status.RepoStatus
	plat := instance.Spec.Plat
	// check Status
	addList, delList := checkStatus(repoList, repoListStatue)
	if plat == "gitlab" {
		webhookToken := generateHmac(secretName)
		statusList := instance.Status.RepoStatus
		if len(addList) != 0 {
			statusList = addMultGitlabBots(statusList, addList, webhookToken)
		}
		if len(delList) != 0 {
			statusList = delMultGitlabBots(statusList, delList)
		}
		instance.Status.RepoStatus = statusList
		secret := &corev1.Secret{}
		err := r.client.Get(context.TODO(), types.NamespacedName{Name: secretName, Namespace: request.Namespace}, secret)
		if err != nil {
			return forget, err
		}
		// update Secret
		secret.Data["webhookToken"] = []byte(webhookToken)
		log.Info("update webhook token", "token", webhookToken)
		err = r.client.Update(context.TODO(), secret)
		if err != nil {
			return forget, err
		}
	}

	// update Status
	err = r.client.Status().Update(context.TODO(), instance.DeepCopy())
	if err != nil {
		log.V(-1).Info("update Failure!", "msg", err)
		return forget, err
	}

	return forget, nil
}

func checkStatus(repoList []string, repoListStatue []maoxianv1.RepoStatus) ([]string, []string) {
	status := set.New(set.ThreadSafe)
	for _, repoObj := range repoListStatue {
		status.Add(repoObj.Name)
	}
	spec := set.New(set.ThreadSafe)
	for _, repo := range repoList {
		spec.Add(repo)
	}
	addList := set.StringSlice(set.Difference(spec, status))
	delList := set.StringSlice(set.Difference(status, spec))
	return addList, delList
}
