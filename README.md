`minikube start --vm-driver=hyperv`

`minikube addons enable ingress`

`minikube docker-env`

`docker build -t downloader:latest -f .\infrastructure\docker\downloader\Dockerfile .`

`docker build -t task_manager:latest -f .\infrastructure\docker\server\Dockerfile .`

`minikube image load task_manager:latest`

`minikube image load downloader:latest`