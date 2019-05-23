build:
	docker build -f docker/Dockerfile -t vothanhkiet/http-proxy .

upload: 
	docker push vothanhkiet/http-proxy 