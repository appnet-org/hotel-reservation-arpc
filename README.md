# Hotel Reservation

The application implements a hotel reservation service, build with Go and aRPC, and starting from the open-source project https://github.com/harlow/go-micro-services. The initial project is extended in several ways, including adding back-end in-memory and persistent databases, adding a recommender system for obtaining hotel recommendations, and adding the functionality to place a hotel reservation. 

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

### Test Application

```bash
curl "http://10.96.88.88:11000/recommendations?require=rate&lat=38.0235&lon=-122.095"
curl "http://10.96.88.88:11000/hotels?inDate=2015-04-10&outDate=2015-04-11&lat=38.0235&lon=-122.095"
curl "http://10.96.88.88:11000/user?username=Cornell_15&password=123654"
curl "http://10.96.88.88:11000/reservation?inDate=2015-04-19&outDate=2015-04-24&lat=nil&lon=nil&hotelId=9&customerName=Cornell_1&username=Cornell_1&password=1111111111&number=1"
```

## Delete Application
```
kubectl delete all,sa,pvc,pv,envoyfilters --all
```
