FROM alpine:latest

COPY ./main /bin/kk-httpd

RUN chmod +x /bin/kk-httpd

ENV KK_NAME kk.httpd.
ENV KK_ALIAS /kk/
ENV KK_ADDR 127.0.0.1:87

EXPOSE 88

CMD kk-httpd $KK_NAME $KK_ADDR :88 $KK_ALIAS
