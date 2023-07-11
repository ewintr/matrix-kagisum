docker-push:
	docker build . -t matrix-kagisum
	docker tag matrix-kagisum registry.ewintr.nl/matrix-kagisum
	docker push registry.ewintr.nl/matrix-kagisum