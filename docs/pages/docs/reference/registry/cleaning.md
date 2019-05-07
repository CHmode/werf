---
title: Cleaning
sidebar: reference
permalink: docs/reference/registry/cleaning.html
author: Artem Kladov <artem.kladov@flant.com>, Timofey Kirillov <timofey.kirillov@flant.com>
---

Build and push processes create sets of docker layers but don't delete them. As a result, local images storage and docker registry instantly grows and consume more and more space. Interrupted build process leaves stale images. When git branch has been deleted, a set of images which was built for this branch also leave in a docker registry. It is necessary to clean a docker registry periodically. Otherwise, it will be filled with **stale images**.

## Methods of cleaning

Werf has an efficient multi-level images cleaning. There are two methods of images cleaning in werf:

1. Cleaning by policies
2. Manual cleaning

### Cleaning by policies

Cleaning by policies helps to organize automatic periodical cleaning of stale images. It implies regular gradual cleaning of stale images according to cleaning policies. This method is the safest way of cleaning because it doesn't affect your production environment.

The cleaning by policies method includes the steps in the following order:
1. [**Cleanup**](#cleanup) cleans docker registry from stale images according to the cleaning policies.
2. [**Local storage synchronization**](#local-storage-synchronization) performs syncing local docker image storage with docker registry.

A docker registry is the primary source of information about actual and stale images. Therefore, it is essential to clean docker registry on the first step and only then synchronize the content of local storage with the docker registry.

### Manual cleaning

Manual cleaning method assumes one-step cleaning with the complete removal of images from local docker image storage or docker registry. This method doesn't check whether the image used by kubernetes or not. Manual cleaning is not recommended for automatical using (use cleaning by policies instead). In general it suitable for forced images removal.

The manual cleaning method includes the following options:

* [**Flush**](#flush). Deleting images of the **current project** in local storage or docker registry.
* [**Reset**](#reset). Deleting images of **all werf projects** in local storage.
* [**GC**](#gc). Force running of werf gc procedure.

## Cleanup

Cleanup is a werf ability to automate cleaning of a docker registry. It works according to special rules called **cleanup policies**. These policies determine which images to delete and which not to.

### Cleanup policies

* **by branches:**
    * Every new commit updates the image for the git branch (there is the only docker tag for the git branch).
    * Werf deletes the image from the docker registry when the corresponding git branch doesn't exist. The image always remains, while the corresponding git branch exists.
    * The policy covers images tagged by werf with `--tag-ci` or `--tag-git-branch` tags.
* **by commits:**
    * Werf deletes the image from the docker registry when the corresponding git commit doesn't exist.
    * For the remaining images, the following policies apply:
       * `WERF_GIT_COMMITS_EXPIRY_DATE_PERIOD_POLICY`. Deleting images uploaded in docker registry more than **30 days**. 30 days is a default period. To change the default period set `WERF_GIT_COMMITS_EXPIRY_DATE_PERIOD_POLICY` environment variable in seconds.
       * `WERF_GIT_COMMITS_LIMIT_POLICY`. Deleting all images in docker registry except **last 50 images**. 50 images is a default value. To change the default value set count in  `WERF_GIT_COMMITS_LIMIT_POLICY` environment variables.
    * The policy covers images tagged by werf with `--tag-git-commit` tag.
* **by tags:**
    * Werf deletes the image from the docker registry when the corresponding git tag doesn't exist.
    * For the remaining images, the following policies apply:
      * `WERF_GIT_TAGS_EXPIRY_DATE_PERIOD_POLICY`. Deleting images uploaded in docker registry more than **30 days**. 30 days is a default period. To change the default period set `WERF_GIT_TAGS_EXPIRY_DATE_PERIOD_POLICY` environment variable in seconds;
      * `WERF_GIT_TAGS_LIMIT_POLICY`.  Deleting all images in docker registry except **last 10 images**. 10 images is a default value. To change the default value set count in `WERF_GIT_TAGS_LIMIT_POLICY`.
    * The policy covers images tagged by werf with `--tag-ci` tag.

**Pay attention,** that cleanup affects only images built by werf **and** images tagged by werf with one of the `--tag-ci`, `--tag-git-branch` or `--tag-git-commit` options. Other images in the docker registry stay as they are.

### Whitelist of images

The image always remains in docker registry while exists kubernetes object which uses the image. In kubernetes cluster werf scans the following kinds of objects: `pod`, `deployment`, `replicaset`, `statefulset`, `daemonset`, `job`, `cronjob`, `replicationcontroller`.

The functionality can be disabled by option `--without-kube`.

#### Connecting to kubernetes

Werf gets information about kubernetes clusters and how to connect to them from the kube configuration file `~/.kube/config`. Werf connects to all kubernetes clusters, defined in all contexts of kubectl configuration, to gather images that are in use.

### Docker registry authorization

For docker registry authorization in cleanup, werf require the `WERF_IMAGES_CLEANUP_PASSWORD` environment variable with access token in it (read more about [authorization]({{ site.baseurl }}/reference/registry/authorization.html#autologin-for-cleaning-commands)).

### Cleanup command

{% include /cli/werf_cleanup.md header="####" %}

## Local storage synchronization

After cleaning docker registry on the cleanup step, local storage still contains all of the images that have been deleted from the docker registry. These images include tagged image images (created with a [werf tag procedure]({{ site.baseurl }}/reference/registry/image_naming.html#werf-tag-procedure)) and stages cache for images.

Executing a local storage synchronization is necessary to update local storage according to a docker registry. On the local storage synchronization step, werf deletes old local stages cache, which doesn't relate to any of the images currently existing in the docker registry.

There are some consequences of this algorithm:

1. If the cleanup, — the first step of cleaning by policies, — was skipped, then local storage synchronization makes no sense.
2. Werf completely removes local stages cache for the built images, that don't exist into the docker registry.

### Sync command

{% include /cli/werf_stages_cleanup.md header="####" %}

## Flush

Allows deleting information about specified (current) project. Flush includes cleaning both the local storage and docker registry.

Local storage cleaning includes:
* Deleting stages cache images of the project. Also deleting images from previous werf version.
* Deleting all of the images tagged by werf with custom tags.
* Deleting `<none>` images of the project. `<none>` images can remain as a result of build process interruption. In this case, such built images exist as orphans outside of the stages cache. These images also not deleted by the cleanup.
* Deleting containers associated with the images of the project.

Docker registry cleaning includes:
* Deleting pushed images of the project.
* Deleting pushed stages cache of the project.

### Flush command

{% include /cli/werf_purge.md header="####" %}

## Reset

With this variant of cleaning, werf deletes all images, containers, and files from all projects created by werf on the host. The files include:
* `~/.werf/{builds|git|worktree|tmp}` directories;
* all lost tmp-dirs generated by werf during builds

Reset is the fullest method of cleaning on the local machine.

### Reset command

{% include /cli/werf_host_purge.md header="####" %}

## GC

Werf has garbage collection procedure, which deletes unused tmp dirs created by werf during its operation.

Normally werf will automatically implicitly run gc procedure while run some command.

There is also gc command to force running of werf gc procedure.

### GC command

{% include /cli/werf_host_cleanup.md header="####" %}
