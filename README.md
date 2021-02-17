
![logo](./assets/images/fotaas-logo.png)

# Formula One Telemetry And Analysis System

### Powered By

<p align="middle">
    <img src="./assets/images/go-logo-2.jpg" width="115" align="center" hspace="10">
    <img src="./assets/images/grpc-logo.png" width="115" align="center" hspace="10">
    <img src="./assets/images/protobuf-logo.png" width="115" align="center" hspace="10">    
    <img src="./assets/images/microservices-logo.jpg" width="115" align="center" hspace="10">
    <img src="./assets/images/docker-logo.png" width="115" align="center" hspace="10">
    <img src="./assets/images/kubernetes-logo.png" width="150" align="center" hspace="10">
    <img src="./assets/images/gcp-logo.png" width="150" align="center" hspace="10">
</p>

## Table of Contents

- [Author](#author)
- [About](#about)
- [Installation](#installation)
- [Usage](#usage)

## Author
Barry T. Burch<br>

Barry is a digital native with over 20 years of experience in software/hardware design and engineering at:

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

* **Golang**
* **Golang Concurrency/Parallelism**
* **Golang Code Generation**
* **Micro-Service Architecure**
* **Protobuf & gRPC**
* **Cobra**
* **Golang Web Application Development**
* **Docker**
* **Kubernetes**
* **Cloud Deployment To GCP (Google Cloud Platform) GKE (Google Kubernetes Environment)**

The FOTAAS system consists of 4 micro-services (telemetry, simulation, analysis, status), a CLI (Command Line Interface)
application, and a Console Web application. The 4 micro-services are completely de-coupled from each other via gRPC APIs
and each service encapsulates a private datastore that can only be accessed via API calls to the service (i.e. this is a
true micro-services based architecture).

## Build & Deployment

The 4 FOTAAS services, console web application, and CLI application are built with docker compose. The resulting docker
images are pushed to GCR (Google Container Registry) and the system is deployed to a GKE (Google Kubernetes Environment)
cluster via kubectl and a non-trivial (i.e. production quality) orchestration yaml.

While you can easily enough clone the FOTASS repo for code review, deploying it to a GKE cluster will not be trivial
(e.g. the GCP Cloud SQL databases would need to created and migrated and this process is not currently documented).

## Usage

To see the FOTAAS system in action please contact me (barry@sbcglobal.net). We can schedule a Google Meet
(or Hangout) and I can demonstrate the FOTASS GCP deployment, CLI usage, and Console Web Application.
