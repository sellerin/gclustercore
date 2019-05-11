// Note: the example only works with the code within the same release/branch.
package gclustercore

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	batchv1 "k8s.io/api/batch/v1"
	//appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	//"k8s.io/client-go/util/retry"
	// Uncomment the following line to load the gcp plugin (only required to authenticate against GKE clusters).
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

func test() {
	fmt.Println("test")
}

func main() {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	jobsClient := clientset.BatchV1().Jobs(apiv1.NamespaceDefault)
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name: "batch-job",
			Namespace: "default",
		},
		Spec: batchv1.JobSpec{
			Parallelism: &(&struct{ x int32 }{2}).x,
			Completions: &(&struct{ x int32 }{2}).x,
			BackoffLimit: &(&struct{ x int32 }{1}).x,
			ActiveDeadlineSeconds: &(&struct{ x int64 }{400}).x,
			Template: apiv1.PodTemplateSpec{
				Spec: apiv1.PodSpec{
					Volumes: []apiv1.Volume{
						{
							Name:  "nfs",
							VolumeSource: apiv1.VolumeSource{
								PersistentVolumeClaim: &apiv1.PersistentVolumeClaimVolumeSource{
									ClaimName: "nfs",
		                        },
							},
						},
						{
							Name:  "user-files",
							VolumeSource: apiv1.VolumeSource{
								GitRepo: &apiv1.GitRepoVolumeSource{
									Repository: "https://github.com/sellerin/gatling-cluster.git",
									Revision: "master",
		                        },
							},
						},
					},
					InitContainers: []apiv1.Container{
						{
							Name:  "prepare-test",
							Image: "busybox",
							Command: []string{"sh", "-c", "rm -rf /exports/results/*; mkdir -p /exports/results;"},
							VolumeMounts: []apiv1.VolumeMount{
								{
									Name: "nfs",
									MountPath: "/results",
								},
							},
						},
					},
					Containers: []apiv1.Container{
						{
							Name:  "main",
							Image: "eu.gcr.io/iron-inkwell-205415/perf:latest",
							Env: []apiv1.EnvVar{
								{
									Name: "SIMULATION_NAME",
									Value: "c2gwebaws.C2gwebSimulation",
								},
							},
							VolumeMounts: []apiv1.VolumeMount{
								{
									Name: "nfs",
									MountPath: "/gatling-charts-highcharts-bundle-3.0.2/results",
									SubPath: "results",
								},
								{
									Name: "user-files",
									MountPath: "/gatling-charts-highcharts-bundle-3.0.2/user-files",
									ReadOnly: true,
									SubPath: "gatling-cluster/user-files",
								},
							},
						},
					},
					RestartPolicy: "Never",
				},
			},
		},
	}

	// Create Job
	fmt.Println("Creating job...")
	job_result, job_err := jobsClient.Create(job)
	if job_err != nil {
		panic(job_err)
	}
	fmt.Printf("Created job %q.\n", job_result.GetObjectMeta().GetName())

	job_watcher := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name: "batch-watcher",
			Namespace: "default",
		},
		Spec: batchv1.JobSpec{
			Parallelism: &(&struct{ x int32 }{1}).x,
			Completions: &(&struct{ x int32 }{1}).x,
			BackoffLimit: &(&struct{ x int32 }{0}).x,
			ActiveDeadlineSeconds: &(&struct{ x int64 }{40000}).x,
			Template: apiv1.PodTemplateSpec{
				Spec: apiv1.PodSpec{
					Volumes: []apiv1.Volume{
						{
							Name:  "nfs",
							VolumeSource: apiv1.VolumeSource{
								PersistentVolumeClaim: &apiv1.PersistentVolumeClaimVolumeSource{
									ClaimName: "nfs",
		                        },
							},
						},
					},
					Containers: []apiv1.Container{
						{
							Name:  "watcher",
							Image: "eu.gcr.io/iron-inkwell-205415/watcher:latest",
							VolumeMounts: []apiv1.VolumeMount{
								{
									Name: "nfs",
									MountPath: "/results",
									SubPath: "results",
								},
								{
									Name: "nfs",
									MountPath: "/aggregated-reports",
								},
							},
						},
					},
					RestartPolicy: "Never",
				},
			},
		},
	}

	// Create Job Watcher
	fmt.Println("Creating job watcher...")
	job_watcher_result, job_watcher_err := jobsClient.Create(job_watcher)
	if job_watcher_err != nil {
		panic(job_watcher_err)
	}
	fmt.Printf("Created job watcher %q.\n", job_watcher_result.GetObjectMeta().GetName())

	/*
	deploymentsClient := clientset.AppsV1().Deployments(apiv1.NamespaceDefault)

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name: "demo-deployment",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(2),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "demo",
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "demo",
					},
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:  "web",
							Image: "nginx:1.12",
							Ports: []apiv1.ContainerPort{
								{
									Name:          "http",
									Protocol:      apiv1.ProtocolTCP,
									ContainerPort: 80,
								},
							},
						},
					},
				},
			},
		},
	}

	// Create Deployment
	fmt.Println("Creating deployment...")
	result, err := deploymentsClient.Create(deployment)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Created deployment %q.\n", result.GetObjectMeta().GetName())

	// Update Deployment
	prompt()
	fmt.Println("Updating deployment...")
	//    You have two options to Update() this Deployment:
	//
	//    1. Modify the "deployment" variable and call: Update(deployment).
	//       This works like the "kubectl replace" command and it overwrites/loses changes
	//       made by other clients between you Create() and Update() the object.
	//    2. Modify the "result" returned by Get() and retry Update(result) until
	//       you no longer get a conflict error. This way, you can preserve changes made
	//       by other clients between Create() and Update(). This is implemented below
	//			 using the retry utility package included with client-go. (RECOMMENDED)
	//
	// More Info:
	// https://github.com/kubernetes/community/blob/master/contributors/devel/api-conventions.md#concurrency-control-and-consistency

	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// Retrieve the latest version of Deployment before attempting update
		// RetryOnConflict uses exponential backoff to avoid exhausting the apiserver
		result, getErr := deploymentsClient.Get("demo-deployment", metav1.GetOptions{})
		if getErr != nil {
			panic(fmt.Errorf("Failed to get latest version of Deployment: %v", getErr))
		}

		result.Spec.Replicas = int32Ptr(1)                           // reduce replica count
		result.Spec.Template.Spec.Containers[0].Image = "nginx:1.13" // change nginx version
		_, updateErr := deploymentsClient.Update(result)
		return updateErr
	})
	if retryErr != nil {
		panic(fmt.Errorf("Update failed: %v", retryErr))
	}
	fmt.Println("Updated deployment...")

	// List Deployments
	prompt()
	fmt.Printf("Listing deployments in namespace %q:\n", apiv1.NamespaceDefault)
	list, err := deploymentsClient.List(metav1.ListOptions{})
	if err != nil {
		panic(err)
	}
	for _, d := range list.Items {
		fmt.Printf(" * %s (%d replicas)\n", d.Name, *d.Spec.Replicas)
	}

	// Delete Deployment
	prompt()
	fmt.Println("Deleting deployment...")
	deletePolicy := metav1.DeletePropagationForeground
	if err := deploymentsClient.Delete("demo-deployment", &metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}); err != nil {
		panic(err)
	}
	fmt.Println("Deleted deployment.")
*/

}

func prompt() {
	fmt.Printf("-> Press Return key to continue.")
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		break
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}
	fmt.Println()
}

func int32Ptr(i int32) *int32 { return &i }