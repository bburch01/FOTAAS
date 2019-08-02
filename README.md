# FOTAAS
Formula One Telemetry Aggregation/Analysis Service

FOTAAS is a portfolio project designed to demonstrate to prospective employers the following proficiencies:

> Golang (including concurrency).
> Micro-Services architecure.
> Protobuf & GRPC.
> Docker.
> Kubernetes.
> Cloud Deployments to GKE (Google Kubernetes Environment).

The FOTAAS system consists of 4 micro-services: telemetry, simulation, analysis, and status. The 4 services
are completely de-coupled from each other via GRPC APIs and each service encapsulates a private datastore
that can only be accessed via API calls to the service (i.e. this is a true micro-services based architecture).

FOTASS includes both a Cobra based CLI (fotaasctl) and a Buffalo powered web application. Both can be used
to exercise the system.

While you an easily enough clone the FOTASS repo for code review, deploying it with Kubernetes will not be
trival (e.g. the Cloud SQL databases need to created and migrated and this process is not currently documented).

If you want to see the FOTAAS Cloud depolyment in action, you will need to contact me (barry@sbcglobal.net). We
can schedule a Google Meet/Hangout and I can demonstrate the GKE deployment for you.

I'm making FOTAAS public (8/2/19) but there is still clean-up work to do:

> Document exported functions.
> Fix broken unit tests.
> Refactor model unit tests to use sqlmock to allow unit tests to run in a CI/CD pipeline.
> Add additional functionality to the FOTAAS web application.




