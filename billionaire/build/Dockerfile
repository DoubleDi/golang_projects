FROM golang:1.17 AS build
WORKDIR /billionaire
ADD go.mod /billionaire
ADD go.sum /billionaire
RUN go mod download

ADD . /billionaire
RUN make tools
RUN make build

FROM debian:buster-slim
WORKDIR /app
COPY --from=build /billionaire/bin/* /app/
COPY --from=build /billionaire/build/*.sh /app/
COPY --from=build /billionaire/deployments/migrations /app/migrations
COPY --from=build /go/bin/migrate /app/
RUN chmod +x /app/*
EXPOSE 8080/tcp
ENTRYPOINT ./entrypoint.sh