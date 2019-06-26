# Industrial Capacity Conversion

Let's say you've got a LOT of projects to complete.  Like over 20,000 (like us).  Many of which run to the hundreds of Mb.

And let's say you've actually got plenty of hardware - in the form of a Kubernetes (or OpenShift) cluster.  Like us.

The kind of thing where taking another hour to set it up is going to save you a day of running time.

Well, you're not going to want to hang around waiting for it all to run on a laptop, are you?

## Step 1 - docker image

You can either use the image available on Docker Hub at `andyg42/premconverter` or you can build your own (if you have Go and make installed)
by running the `build_docker.sh` script in the root of the repo.  To run in Kubernetes, you'll need to push it to a repo
somewhere - either Docker Hub, your own one, or whatever.

## Step 2 - get the file lists and decide on parallelism

I ran `find $PWD -iname \*.prproj > ~/allprojects.lst` in the directory where we keep our projects to get a list of the projects.
You'll want the whole folder path in there, hence using `$PWD` instead of `.`.

While the app allows for concurrency, that is within one instance. To effectively spread over hardware, we also want
to have multiple instances going.  Say you want 5 instances, then run:

`./split-out-list.py --input allprojects.lst --output-base my-projects --parts 5` in this directory
to generate 5 files called my-projects-{{n}}.lst.

_Why is that Python_? I hear you ask - well it's because I already had it lying around from another project. If that is
a big problem it shouldn't be hard to reimplement in Go.

## Step 3 - provision storage

You need to get these list files into the Kubernetes cluster somehow.  In our cluster, the simplest way is to provision
some NFS storage and copy them there using kubectl onto a dummy pod:

```console
$ kubectl apply -f lists-storage.yaml
$ kubectl apply -f provisioning-pod.yaml
$ kubectl exec -it provisioning-pod /bin/sh
[wait....]
/ $ exit
$ for x in `ls my-projects-*.lst`; do kubectl cp $x provisioning-pod:/mnt/lists; done
$ kubectl exec -it provisioning-pod /bin/sh
/ $ ls /mnt/lists/
my-projects-1.lst  my-projects-3.lst  my-projects-5.lst
my-projects-2.lst  my-projects-4.lst
/ $ exit
$ kubectl delete -f provisioning-pod.yaml
```

You may need to tweak the manifests to provision whatever kind of storage is relevant for your clusters.

## Step 4 - set up jobs

In our system, our project files live on a shared SAN that some of our Kubernetes nodes have access to (denoted by the
`has-san=true` label).  So we can simply bind-mount that from the local path of the node.

You will probably need to adjust the `templates/RunAsJob.yaml` file:
 
- Ensure that the `input-path` and `output-path` volumes are set up as required for your setup. If you are able to bind-mount like us,
 the template should be fine as-is.  Otherwise you'll need to make your projects available to the cluster and provide an output path as well.
- The template contains a node selector (under `spec.node.selector`) to only run on nodes where the label `has-san=true` exists; you will need to remove this or
update to whatever is relevant for your setup
- The template also contains an anti-affinity to try to spread jobs across all available nodes.  This is good if you want to spread the
workload across your cluster, but you might want to change this if you want to keep the jobs away from some of your nodes.
- Finally, the template contains resource requests again to try to stop the system from placing _all_ the pods on the same node.

Next, edit the `jobs-from-template.sh` script and adjust the settings at the top:

```
NUM_SPLITS={number-of-splits-frompart2}
REAL_INPUT_PATH=/path-to-bindmount
REAL_OUTPUT_PATH=/path-to-bindmount
LISTFILE=dev-environment-projects
```

This should create a collection of job manifests called `premconverter-job-{n}.yaml`, each of which points to one of the
manifest files you generated in steps 2 and 3.
Once you've checked that the volume configuration works, you should just be able to apply them:

```console
$ for x in `ls premconverter-job-*`; do kubectl apply -f $x; done
job.batch/premiere-converter-1 created
job.batch/premiere-converter-2 created
job.batch/premiere-converter-3 created
job.batch/premiere-converter-4 created
job.batch/premiere-converter-5 created
$
``` 

Once the jobs are set up, you can monitor each however you normally monitor a pod (from the commandline, `kubectl logs -f job.batch/premiere-converter-1` etc.)

Remember, if you make a change and then want to re-run you'll need to delete the jobs to start over:
```console
$ for x in `ls premconverter-job-*`; do kubectl delete -f $x; done
job.batch "premiere-converter-1" deleted
job.batch "premiere-converter-2" deleted
job.batch "premiere-converter-3" deleted
job.batch "premiere-converter-4" deleted
job.batch "premiere-converter-5" deleted
```

## Considerations

### Overwriting behaviour
Over-writing of already existing files in the output directory is disabled by default.  You'll see a message in the log if a conversion
is blocked because of this and it'll count as "failed" in the final stats you see at the end of the run.
If you want to change this, add "--allow-overwrite" to the end of the arguments list in `templates/RunAsJob.yaml` around line 30 (containers.command)

