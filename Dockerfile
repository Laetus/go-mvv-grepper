FROM golang:latest 
RUN go get -u github.com/golang/dep/cmd/dep
RUN mkdir /go/src/app
ADD ./main.go /go/src/app
COPY ./Gopkg.toml /go/src/app
WORKDIR /go/src/app 
RUN dep ensure 
RUN go test -v 
RUN go build main.go
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .


FROM alpine:latest  
LABEL version="1.0"
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=0 /go/src/app .
CMD ["./main"]  
