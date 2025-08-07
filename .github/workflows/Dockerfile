FROM ubuntu:latest

WORKDIR /app

RUN apt-get update && apt-get install -y golang
RUN apt-get install -y ca-certificates

COPY . .

RUN go mod tidy
RUN find caisse_app_scaled -type f -exec sed -i 's|^\(\s*_ "caisse-app-scaled/docs/swagger/[^"]*"\)|// \1|' {} + ||true

CMD ["go", "test", "./..."]
