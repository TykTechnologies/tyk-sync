#Docker
To build a docker image for `tyk-sync`, run:
```
docker build -t tyk-sync:latest .
```
To run `tyk-sync` run
```
docker run -it --network="host" -v $(pwd)/apis:/apis --rm tyk-sync:latest tyk-sync dump -s <admin_secret> -d https://tyk-dashboard.test-backend.vdbinfra.nl/ -t /apis
```
this will dump api definitions into the `./apis` folder
