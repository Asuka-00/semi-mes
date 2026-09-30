FROM alpine:latest
MAINTAINER zhufuyi "g.zhufuyi@gmail.com"

# add curl, used for http service checking, can be installed without it if deployed in k8s
RUN apk add curl

COPY configs/ /app/configs/
COPY mes /app/mes
RUN chmod +x /app/mes

# http port
EXPOSE 8080


WORKDIR /app

CMD ["./mes", "-c", "configs/mes.yml"]
# if you use the Configuration Center, mes.yml is changed to the Configuration Center configuration.
#CMD ["./mes", "-c", "configs/mes.yml", "-enable-cc"]
