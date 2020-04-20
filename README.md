# FOTAAS
<p align="center">
    <img src="./assets/images/fotass-logo.png">
</p>

<p align="center">
    Formula One Telemetry And Analysis System: A Golang micro-services based system simulates the collection, persistence, and analyis of F1 race telemetry.
</p>

<p align="center">
    <img src="./assets/images/go-logo.png" height="100" width="100">
</p>

## Table of Contents

- [Author](#author)
- [About](#about)
- [Installation](#installation)
- [Usage](#usage)
- [UnitTest](#unittest)
- [Deficiencies](#deficiencies)

## Author
Barry T. Burch<br>

Barry is a digital native with over 20 years of experience in software (and hardware) design and engineering at:

<p align="middle">
    <img src="./assets/images/ti-logo-2.png" align="center" hspace="10">
    <img src="./assets/images/nec-logo-2.png" align="center" hspace="10">
    <img src="./assets/images/att-logo-2.jpeg" align="center" hspace="20">
    <img src="./assets/images/avaya-logo-2.png" width="100" align="center" hspace="10">
    <img src="./assets/images/sxm-logo.jpeg" width="100" align="center" hspace="10">
    <img src="./assets/images/gf-logo.jpeg" width="100" align="center" hspace="10">
</p>

barry@sbcglobal.net<br>
www.linkedin.com/in/barry-burch-digital-native<br>

## About

FOTAAS is a Golang portfolio project designed to demonstrate the following proficiencies:

> Golang
> Golang Concurrency/Parallelism
> True Micro-Services Architecure
> Protobuf & gRPC
> Golang Web Application Development
> Docker
> Kubernetes
> Cloud Deployment To GCP (Google Cloud Platform) GKE (Google Kubernetes Environment)
> GIT
> MOD

The FOTAAS system consists of 4 micro-services: telemetry, simulation, analysis, and status. The 4 services
are completely de-coupled from each other via GRPC APIs and each service encapsulates a private datastore
that can only be accessed via API calls to the service (i.e. this is a true micro-services based architecture).

FOTASS includes both a Cobra based CLI (fotaasctl) and a console web application. Both can be used
to exercise the system.

While you an easily enough clone the FOTASS repo for code review, deploying it with Kubernetes will not be
trival (e.g. the Cloud SQL databases would need to created and migrated and this process is not currently documented).

If you want to see the FOTAAS Cloud depolyment in action, you will need to contact me (barry@sbcglobal.net). We
can schedule a Google Meet/Hangout and I can demonstrate the GKE deployment for you.

## Installation

## Usage

## UnitTest

## Deficiencies





