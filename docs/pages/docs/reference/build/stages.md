---
title: Stages
sidebar: reference
permalink: docs/reference/build/stages.html
author: Alexey Igrychev <alexey.igrychev@flant.com>
---

What usually needs for build application image?

* Choose a base image
* Add source code
* Install system dependencies and software
* Install application dependencies
* Configure system software
* Configure application

In what order do you need to perform these steps for the effective assembly (re-assembly) process?

We propose to divide the assembly into steps with clear functions and purposes. In werf, such steps are called _stages_.

## What is a stage?

A ***stage*** is a logically grouped set of config instructions, as well as the conditions and rules by which these instructions are assembled.

The werf assembly process is a sequential build of _stages_. Werf uses different _stage conveyor_ for assembling a particular type of build object. A ***stage conveyor*** is a statically defined sequence of _stages_. The set of _stages_ and their order is predetermined.

<div class="tab">
  <button class="tablinks active" onclick="openTab(event, 'image')">Image</button>
  <button class="tablinks" onclick="openTab(event, 'artifact')">Artifact</button>
</div>

<div id="image" class="tabcontent active">
<a class="google-drawings" href="https://docs.google.com/drawings/d/e/2PACX-1vRKB-_Re-ZhkUSB45jF9GcM-3gnE2snMjTOEIQZSyXUniNHKK-eCQl8jw3tHFF-a6JLAr2sV73lGAdw/pub?w=2000&amp;h=881" data-featherlight="image">
<img src="https://docs.google.com/drawings/d/e/2PACX-1vRKB-_Re-ZhkUSB45jF9GcM-3gnE2snMjTOEIQZSyXUniNHKK-eCQl8jw3tHFF-a6JLAr2sV73lGAdw/pub?w=821&amp;h=362" >
</a>
</div>

<div id="artifact" class="tabcontent">
<a class="google-drawings" href="https://docs.google.com/drawings/d/e/2PACX-1vRD-K_z7KEoliEVT4GpTekCkeaFMbSPWZpZkyTDms4XLeJAWEnnj4EeAxsdwnU3OtSW_vuKxDaaFLgD/pub?w=1800&amp;h=850" data-featherlight="image">
<img src="https://docs.google.com/drawings/d/e/2PACX-1vRD-K_z7KEoliEVT4GpTekCkeaFMbSPWZpZkyTDms4XLeJAWEnnj4EeAxsdwnU3OtSW_vuKxDaaFLgD/pub?w=640&amp;h=301">
</a>
</div>

<!-- 301 -->

**All works with _stages_ are done by werf, and you only need to write config correctly.**

Each _stage_ is assembled in an ***assembly container*** based on an image of the previous _stage_. The result of the assembly _stage_ and _stage conveyor_, in general, is the ***stages cache***: each _stage_ relates to one docker image.

Using a cache for re-assemblies is possible due to the build stage identifier called _signature_. The _signature_ is calculated for the _stages_ at each build. At the last step of the build when saving _stages cache_, the _signature_ is used for tagging (`image-stage-<project name>:<signature>`). This logic allows to assembly only _stages_ whose the _stages cache_ does not exist in the docker.

<div class="rsc" markdown="1">

<div class="rsc-description" markdown="1">

  The ***stage signature*** is the checksum of _stage dependencies_ and previous _stage signature_. In the absence of _stage dependencies_, the _stage_ is skipped.

  It means that the _stage conveyor_, e.g., image _stage conveyor_, can be reduced to several _stages_ or even to single _from_ stage.

</div>

<div class="rsc-example">
<a class="google-drawings" href="https://docs.google.com/drawings/d/e/2PACX-1vR6qxP5dbQNlHXik0jCvEcKZS2gKbdNmbFa8XIem8pixSHSGvmL1n7rpuuQv64YWl48wLXfpwbLQEG_/pub?w=572&amp;h=577" data-featherlight="image">
<img src="https://docs.google.com/drawings/d/e/2PACX-1vR6qxP5dbQNlHXik0jCvEcKZS2gKbdNmbFa8XIem8pixSHSGvmL1n7rpuuQv64YWl48wLXfpwbLQEG_/pub?w=286&amp;h=288">
</a>
</div>

</div>

<div style="clear: both;"></div>
