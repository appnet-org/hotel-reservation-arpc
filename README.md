# Hotel Reservation

The application implements a hotel reservation service, build with Go and aRPC, and starting from the open-source project https://github.com/harlow/go-micro-services. The initial project is extended in several ways, including adding back-end in-memory and persistent databases, adding a recommender system for obtaining hotel recommendations, and adding the functionality to place a hotel reservation. 

<!-- ## Application Structure -->

<!-- ![Social Network Architecture](socialNet_arch.png) -->

Supported actions: 
* Get profile and rates of nearby hotels available during given time periods
* Recommend hotels based on user provided metrics
* Place reservations

## Build Kubernetes images

```bash
git clone https://github.com/appnet-org/go-lib.git
bash kubernetes/scripts/build-docker-images.sh
```

## Run Application
```bash
kubectl apply -f hotel_reservation.yaml
```


## Delete Application
```
kubectl delete all,sa,pvc,pv,envoyfilters,appnetconfigs --all
```
