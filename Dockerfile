# создаем образ на основе Alpine Linux, назывыем builder
FROM golang:1.23-alpine AS builder

# устанавливаем рабочую дирректорию "по канонам"(по сути мы перешли по этой дирректории)
WORKDIR /usr/local/src

# устанавливаем оболочку bash, git в образе, не кэшируя данные
RUN apk --no-cache add bash git gcc musl-dev

# устанавливаем зависимсоти, первые пареметры - что копируем, последний параметр - куда
# (в нашем случае в ту директорию, в которой находимся - "./", а текущая дерриктория - 
#                                                                  /usr/local/src)
COPY ["go.mod", "go.sum", "fs.go", "./"]

# скачивает все зависимости, указанные в файле go.mod
RUN go mod download

# docker build -t web_app2:v1 .  - запускаем команду в терминали -t(tag) флаг, котороый дает 
# имя и тег нашему образу web_app2 - имя, v1 - тег 

# копируем все файлы из директории (cmd,internal,static.README.md в дирректорию докер, 
# в нашем случае в /usr/local/src (первый аргумент это наш код, второй - директория
# в образе докера
COPY ./ ./

# билдим
RUN go build -o ./bin/app ./cmd/main.go

FROM alpine AS runner

COPY --from=builder /usr/local/src/bin/app / 

CMD ["/app"]

#docker run -d --name container1  -p 8080:8080 web_app2:v1 - запуск контйнера после создания образа


