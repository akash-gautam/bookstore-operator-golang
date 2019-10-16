package bookstore

import (
	"context"
  "reflect"
	blogv1alpha1 "bookstore-operator/pkg/apis/blog/v1alpha1"
  "k8s.io/apimachinery/pkg/api/resource"
	corev1 "k8s.io/api/core/v1"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_bookstore")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new BookStore Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileBookStore{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("bookstore-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource BookStore
	err = c.Watch(&source.Kind{Type: &blogv1alpha1.BookStore{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}
	return nil
}

// blank assignment to verify that ReconcileBookStore implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileBookStore{}

// ReconcileBookStore reconciles a BookStore object
type ReconcileBookStore struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a BookStore object and makes changes based on the state read
// and what is in the BookStore.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileBookStore) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling BookStore")

	// Fetch the BookStore instance
	instance := &blogv1alpha1.BookStore{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		reqLogger.Error(err, "Failed to get bookstore object")
		return reconcile.Result{}, err
	}
	err = r.BookStore(instance)
	if err != nil {
			reqLogger.Error(err, "Failed to create/update bookstore resources")
			return reconcile.Result{}, err
	}
	_ = r.client.Status().Update(context.TODO(), instance)
	return reconcile.Result{}, nil
}


func (r *ReconcileBookStore) BookStore(bookstore *blogv1alpha1.BookStore) error {
	reqLogger := log.WithValues("Namespace", bookstore.Namespace)
	mongoDBSvc := getmongoDBSvc(bookstore)
	msvc := &corev1.Service{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: "mongodb-service", Namespace: bookstore.Namespace}, msvc)
	if err != nil {
		if errors.IsNotFound(err) {
			controllerutil.SetControllerReference(bookstore, mongoDBSvc, r.scheme)
	    err = r.client.Create(context.TODO(), mongoDBSvc)
	    if err != nil { return err }
		} else {
			   return err
		}
	} else if !reflect.DeepEqual(mongoDBSvc.Spec, msvc.Spec) {
			mongoDBSvc.ObjectMeta = msvc.ObjectMeta
			controllerutil.SetControllerReference(bookstore, mongoDBSvc, r.scheme)
			err = r.client.Update(context.TODO(), mongoDBSvc)
	    if err != nil { return err }
			reqLogger.Info("mongodb-service updated")
		 }
	mongoDBSS := getMongoDBStatefulsets(bookstore)
	mss := &appsv1.StatefulSet{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: "mongodb", Namespace: bookstore.Namespace}, mss)
	if err != nil {
		if errors.IsNotFound(err) {
			reqLogger.Info("mongodb statefulset not found, will be created")
	    controllerutil.SetControllerReference(bookstore, mongoDBSS, r.scheme)
	    err = r.client.Create(context.TODO(), mongoDBSS)
	    if err != nil { return err }
			} else {
				reqLogger.Info("failed to get mongodb statefulset")
				return err
			}
	} else if !reflect.DeepEqual(mongoDBSS.Spec, mss.Spec) {
					r.UpdateVolume(bookstore)
					mongoDBSS.ObjectMeta = mss.ObjectMeta
					mongoDBSS.Spec.VolumeClaimTemplates = mss.Spec.VolumeClaimTemplates
				  controllerutil.SetControllerReference(bookstore, mongoDBSS, r.scheme)
				  err = r.client.Update(context.TODO(), mongoDBSS)
				  if err != nil { return err }
				  reqLogger.Info("mongodb statefulset updated")
		}
	bookStoreSvc := getBookStoreAppSvc(bookstore)
	bsvc := &corev1.Service{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: "bookstore-svc", Namespace: bookstore.Namespace}, bsvc)
	if err != nil {
		if errors.IsNotFound(err) {
				controllerutil.SetControllerReference(bookstore, bookStoreSvc, r.scheme)
	      err = r.client.Create(context.TODO(), bookStoreSvc)
	      if err != nil { return err }
			} else {
			reqLogger.Info("failed to get bookstore service")
			return err
		}
	} else if !reflect.DeepEqual(bookStoreSvc.Spec, bsvc.Spec) {
				bookStoreSvc.ObjectMeta = bsvc.ObjectMeta
				bookStoreSvc.Spec.ClusterIP = bsvc.Spec.ClusterIP
			  controllerutil.SetControllerReference(bookstore, bookStoreSvc, r.scheme)
			  err = r.client.Update(context.TODO(), bookStoreSvc)
	      if err != nil { return err }
			  reqLogger.Info("bookstore service updated")
	}
	bookStoreDep := getBookStoreDeploy(bookstore)
	bsdep := &appsv1.Deployment{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: "bookstore", Namespace: bookstore.Namespace}, bsdep)
	if err != nil {
		if errors.IsNotFound(err) {
			  controllerutil.SetControllerReference(bookstore, bookStoreDep, r.scheme)
	      err = r.client.Create(context.TODO(), bookStoreDep)
	      if err != nil { return err }
		} else {
			reqLogger.Info("failed to get bookstore deploymnet")
			return err
		}
	} else if !reflect.DeepEqual(bookStoreDep.Spec, bsdep.Spec) {
				bookStoreDep.ObjectMeta = bsdep.ObjectMeta
			  controllerutil.SetControllerReference(bookstore, bookStoreDep, r.scheme)
			  err = r.client.Update(context.TODO(), bookStoreDep)
	      if err != nil { return err }
			reqLogger.Info("bookstore deployment updated")
		 }
	r.client.Status().Update(context.TODO(), bookstore)
  return nil
}

func getBookStoreAppSvc(bookstore *blogv1alpha1.BookStore) *corev1.Service {

	p := make([]corev1.ServicePort,0)
	servicePort := corev1.ServicePort{
		Name: "tcp-port",
		Port: bookstore.Spec.BookApp.Port,
		TargetPort: intstr.FromInt(bookstore.Spec.BookApp.TargetPort),
	}
	p = append(p, servicePort)
     bookStoreSvc := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        "bookstore-svc",
			Namespace:    bookstore.Namespace,
			Labels:      map[string]string{"app": "bookstore"},
		},
		Spec: corev1.ServiceSpec{
			Ports:     p,
			Type:      bookstore.Spec.BookApp.ServiceType,
			Selector:  map[string]string{"app": "bookstore"},
		},
	}
	return bookStoreSvc
}

func getmongoDBSvc(bookstore *blogv1alpha1.BookStore) *corev1.Service {

	p := make([]corev1.ServicePort,0)
	servicePort := corev1.ServicePort{
		Name: "tcp-port",
		Port: bookstore.Spec.BookDB.Port,
	}
	p = append(p, servicePort)
	mongoDBSvc := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        "mongodb-service",
			Namespace:    bookstore.Namespace,
			Labels:      map[string]string{"app": "bookstore-mongodb"},
		},
		Spec: corev1.ServiceSpec{
			Ports:     p,
			Selector:  map[string]string{"app": "bookstore-mongodb"},
			ClusterIP: "None",
		},
	}
	return mongoDBSvc
}

func getBookStoreDeploy(bookstore *blogv1alpha1.BookStore) *appsv1.Deployment {
	
	cnts := make([]corev1.Container,0)
	cnt := corev1.Container{
		Name:  "bookstore",
		Image: bookstore.Spec.BookApp.Repository+":"+bookstore.Spec.BookApp.Tag,
		ImagePullPolicy: bookstore.Spec.BookApp.ImagePullPolicy,
	}
    cnts = append(cnts, cnt)
	podTempSpec := corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{"app": "bookstore"},
		},
		Spec: corev1.PodSpec{
			Containers: cnts,
		},
	}
	dep := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "bookstore",
			Namespace:  bookstore.Namespace,
			Labels:    map[string]string{"app": "bookstore"},
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": "bookstore"},
			},
			Replicas: &bookstore.Spec.BookApp.Replicas,
			Template: podTempSpec,
		},
	}
	return dep	
}
func getMongoDBStatefulsets(bookstore *blogv1alpha1.BookStore) *appsv1.StatefulSet {

	cnts := make([]corev1.Container,0)
	cnt := corev1.Container{
		Name: "mongodb",
		Image: bookstore.Spec.BookDB.Repository+":"+bookstore.Spec.BookDB.Tag,
		ImagePullPolicy: bookstore.Spec.BookDB.ImagePullPolicy,
	}
    cnts = append(cnts, cnt)
	podTempSpec := corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels:map[string]string{"app": "bookstore-mongodb"},
		},
		Spec: corev1.PodSpec{
			Containers: cnts,
		},
	}
	mongoss := &appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "StatefulSet",
			APIVersion: "v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "mongodb",
			Namespace: bookstore.Namespace,
			Labels:    map[string]string{"app": "bookstore-mongodb"},
		},
		Spec: appsv1.StatefulSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": "bookstore-mongodb"},
			},
			Replicas:       &bookstore.Spec.BookDB.Replicas,
			Template:       podTempSpec,
			ServiceName:    "mongodb-service",
			//VolumeClaimTemplates: volClaimTemplate(),
			VolumeClaimTemplates: volClaimTemplate(bookstore.Spec.BookDB.DBSize),
		},
	}
   return mongoss
}

func volClaimTemplate(DBSize resource.Quantity) []corev1.PersistentVolumeClaim {

  storageClass := "standard"
  mongorr := corev1.ResourceRequirements{
	Requests: corev1.ResourceList{
		//corev1.ResourceStorage: resource.MustParse(DBSize),
		corev1.ResourceStorage: DBSize,
	 },
  }
  accessModeList := make([]corev1.PersistentVolumeAccessMode,0)
  accessModeList = append(accessModeList,corev1.ReadWriteOnce)
  mongopvc := corev1.PersistentVolumeClaim{
  	ObjectMeta: metav1.ObjectMeta{
		Name: "mongodb-pvc",
	},
	Spec: corev1.PersistentVolumeClaimSpec{
		AccessModes:      accessModeList,
		Resources:        mongorr,
		StorageClassName: &storageClass,
	},
   }
  pvcList := make([]corev1.PersistentVolumeClaim,0)
  pvcList = append(pvcList, mongopvc)
  return pvcList
}
func (r *ReconcileBookStore) UpdateVolume(bookstore *blogv1alpha1.BookStore) error {

	reqLogger := log.WithValues("Namespace", bookstore.Namespace)
  mpvc := &corev1.PersistentVolumeClaim{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: "mongodb-pvc-mongodb-0", Namespace: bookstore.Namespace}, mpvc)
	if err != nil {
		return nil
	}
  if mpvc.Spec.Resources.Requests[corev1.ResourceStorage] != bookstore.Spec.BookDB.DBSize {
		reqLogger.Info("Need to expand the mongodb volume")
		mpvc.Spec.Resources.Requests[corev1.ResourceStorage] = bookstore.Spec.BookDB.DBSize
		err := r.client.Update(context.TODO(), mpvc)
		if err != nil {
			reqLogger.Info("Error in expanding the mongodb volume")
			return err
		}
		reqLogger.Info("mongodb volume updated successfully")
	}
  return nil
}