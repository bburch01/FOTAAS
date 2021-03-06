
![logo](./assets/images/fotaas-logo.png)

# Formula One Telemetry And Analysis System

## FOTAAS IS Powered By

<p align="middle">
    <img src="./assets/images/go-logo-2.jpg" width="115" align="center" hspace="10">
    <img src="./assets/images/grpc-logo.png" width="115" align="center" hspace="10">
    <img src="./assets/images/protobuf-logo.png" width="115" align="center" hspace="10">    
    <img src="./assets/images/microservices-logo.jpg" width="115" align="center" hspace="10">
    <img src="./assets/images/docker-logo.png" width="115" align="center" hspace="10">
    <img src="./assets/images/kubernetes-logo.png" width="150" align="center" hspace="10">
    <img src="./assets/images/gcp-logo.png" width="150" align="center" hspace="10">
    <img src="./assets/images/gke-logo-2.png" width="150" align="center" hspace="10">
    <img src="./assets/images/gcsql-logo.png" width="150" align="center" hspace="10">
</p>

## About The Author
Barry T. Burch<br>

Barry is a digital native with over 30 years of experience in software/hardware design and engineering at:

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

## About The FOTAAS Project

FOTAAS is a Golang portfolio project designed to demonstrate technical proficiency with:

* **Golang**
* **Golang Concurrency/Parallelism**
* **Golang Code Generation**
* **Microservice Architecture**
* **Protobuf**
* **gRPC**
* **Cobra**
* **Golang Web Application Development**
* **Docker**
* **Kubernetes**
* **Cloud Deployment**

The FOTAAS deployment consists of 4 microservices (telemetry, simulation, analysis, status), a CLI (Command Line Interface)
application, a Console Web application, and a GCP Cloud SQL database. The 4 micro-services are completely de-coupled from
each other via gRPC APIs and each service encapsulates a private datastore that can only be accessed via API calls to the
that service (i.e. this is a true microservices based architecture).

## FOTAAS Build And Cloud Deployment

The 4 FOTAAS services, console web application, and CLI application are built with docker compose. The resulting
docker images are pushed to GCR (Google Container Registry) and the system is then deployed via kubectl and an
orchestration yaml. The deployment target is a GKE (Google Kubernetes Engine) environment hosted by a GCP
(Google Cloud Platform) Compute Engine cluster comprised of 4 nodes running in the us-central1-b zone.
        
## FOTAAS In Action

To see the FOTAAS system in action please contact me (barry@sbcglobal.net). We can schedule a Google Meet
(or Hangout) and I can demonstrate the FOTAAS GCP deployment, CLI usage, and Console Web Application.
