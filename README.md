# Honey Dipper

<!-- toc -->

- [Overview](#overview)
- [Design](#design)
  * [Core Concepts](#core-concepts)
  * [Features](#features)
    + [Embracing GitOps](#embracing-gitops)
    + [Pluggable Architecture](#pluggable-architecture)
    + [Abstraction](#abstraction)
- [TODO](#todo)
- [Get Started On Developing](#get-started-on-developing)

<!-- tocstop -->

## Overview
A IFTTT style event-driven, policy-based orchestration system that, is tailored towards SREs and DevOps workflows, and has a pluggable open architecture.  The purpose is to fill the gap between the various components used in DevOps operations to act as an orchestration hub, and to replace the ad-hoc integrations between the components so that the all the integrations can also be composed as code.

## Design
The core of the Honey Dipper is comprised of an event bus, and a rules engine.  Raw events from various sources are received by corresponding event drivers, and then packaged in a standard format then published to the event bus.  The rules engine picks up the event from the bus, and, based on the rules, triggers the actions or a workflow. 
![Dipper Architecture](./DipperDiagram1.png)

### Core Concepts
In order for users to compose the rules, a few abstract concept is introduced.

 * Driver (Event)
 * Raw Event
 * System (Trigger): an abstract entity that groups dipper events and some configurations, metadata together 
 * Dipper Event (DipperMessage): a data structure that contains information that can be used for matching rules and being processed following the rules
 * Rules: if some Dipper Event on some system happens, then start the workflow of actions on certain systems accordingly
 * Filters: Functions to be called to mutate the data structure in the event/action so various event/action can be linked together based on a contract
 * Workflow: Grouping of the actions so they can be processed, sequentially, parallel, etc
 * Dipper Action (DipperMessage): a data structure that contains information that can be used for performing an action
 * System (Responder): an abstract entity that groups dipper actions, configurations, metadata together
 * Raw Action
 * Driver (Action)

As you can see, the item described above follows the order or life cycle stage of the processing of the events into actions.  Ideally, anything between the drivers should be composable, while some may tend to focusing on making various systems, dipper event/actions available, others may want to focus on rules, workflows.   

### Features

#### Embracing GitOps
Honey Dipper should have minimum to almost none local configuration to bootstrap.  Once bootstrapped, the system should be able to pull configurations from one or more git repo.  The benefit is the ease of maintenance of the system and access control automatically provided by git repo(s).  The system needs to watch the git repos, one way or another, for changes and reload as needed.  For continuous operation, the system should be able to survive when there is a configure error, and should be able to continue running with an older version of the configuration.

#### Pluggable Architecture
Drivers make up an important part of the Honey Dipper ecosystem.  Most of the data mutation and actual work process are done by the drivers, including data decryption, internal communication, interacting with external systems. Honey Dipper should be able to extend itself through loading external drivers dynamically, and when configurations change, reload the running drivers hot or cold.  There should be an interface for the drivers to delegate work to each other through RPC.

#### Abstraction
As mentioned in the concepts, one of Honey Dipper's main selling points is "abstraction".  Events, actions can be defined traditionally using whatever characteristics provided by a driver, but also can be defined as an extension of another event/action with additional or override parameters.  Events and actions can be grouped together into systems where data can be shared across.  With this abstraction, we can separate the composing of complex workflows from defining low level event/action hook ups.  Whenever a low level component changes, the high level workflow doesn't have to change, just hook the abstract events with the new component native events.

![Dipper Daemon](./DipperDaemon.png)

## TODO
 * Test framework and CI
 * dockerizing
 * kubernetes manifest files
 * Documentation for users
 * Documentation for developers
 * Enhancing the driver communication protocol to support more encodings, gob, pickle
 * Python driver library
 * API service
 * Dashboard webapp
 * Auditing/logging driver
 * State persistent driver
 * Config file templating
 * Data/parameter templating
 * Repo jailing
 * RBAC

## Get Started On Developing

 * Setup a directory as your go work directory and add it to GOPATH. Assuming go 1.11 or up is installed, gvm is recommended to manage multiple versions of go. You may want to persist the GOPATH in your bash_profile
```bash
mkdir gocode
export GOPATH=$GOPATH:$PWD/gocode
```
 * Clone the code
```bash
cd gocode
mkdir -p src/github.com/honeyscience/
cd src/github.com/honeyscience/
git clone git@github.com:honeyscience/honeydipper.git
```
 * Load the dependencies
```bash
brew install dep
cd honeydipper
dep ensure
```
 * Run test
```bash
go test -v
```
 * (Optional) For colored test results
```bash
go get -u github.com/rakyll/gotest
gotest -v
```
 * (Optional) For pre-commit hooks
```bash
brew install pre-commit
pre-commit install --install-hooks
```
 * Start your local dipper daemon, see admin guide(coming soon) for detail
```bash
REPO=file:///path/to/your/rule/repo honeydipper
```