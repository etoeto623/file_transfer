FROM alpine:3.15.0
COPY ft /usr/local/
COPY .ftrc /root/.ftrc
WORKDIR /usr/local/

RUN chmod u+x /usr/local/ft
RUN mkdir /root/warehouse
CMD ["./ft", "serve"]