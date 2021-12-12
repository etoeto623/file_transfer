FROM alpine:3.15.0
COPY ft /usr/local/ft
COPY .ftrc /root/.ftrc
# RUN echo "https://mirror.tuna.tsinghua.edu.cn/alpine/v3.8/main" > /etc/apk/repositories
# RUN apk add --update curl bash && rm -rf /var/cache/apk/*
WORKDIR /usr/local/

RUN chmod u+x /usr/local/ft
RUN mkdir /root/warehouse
CMD ["/usr/local/ft", "serve"]