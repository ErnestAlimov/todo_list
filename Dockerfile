FROM golang:1.20-buster

RUN go version
ENV GOPATH=/

COPY ./ ./

RUN go mod download
RUN go build -o todo_list ./main.go

CMD [ "./todo_list" ]