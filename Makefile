engine:
	go build -o e-wallet ./cmd/.

clean:
	if [ -f e-wallet ]; then rm e-wallet ; fi

image:
	docker build -t e-wallet-image .

run:
	docker run -d --name e-wallet -p 8000:8000 e-wallet-image

stop:
	docker container stop e-wallet

dev:
	go run ./cmd/.
