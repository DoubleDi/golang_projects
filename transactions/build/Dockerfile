FROM golang:1.14 AS build
WORKDIR /transactions
ADD go.mod /transactions
ADD go.sum /transactions
RUN go mod download

ADD . /transactions
RUN make tools
RUN make build

FROM debian:buster-slim
WORKDIR /app
COPY --from=build /transactions/bin/* /app/
COPY --from=build /transactions/build/*.sh /app/
COPY --from=build /transactions/deployments/migrations /app/migrations
COPY --from=build ${HOME}/go/bin/migrate /app/migrate
RUN chmod +x /app/*
EXPOSE 8080/tcp
CMD /app/transactions
