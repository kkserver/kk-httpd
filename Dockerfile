FROM alpine:latest

RUN echo "Asia/shanghai" >> /etc/timezone

COPY ./main /bin/kk-httpd

RUN chmod +x /bin/kk-httpd

COPY ./config /config

COPY ./app.ini /app.ini

ENV KK_ENV_CONFIG /config/env.ini

VOLUME /config

EXPOSE 88

CMD kk-httpd $KK_ENV_CONFIG


