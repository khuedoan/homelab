FROM nginx:1.19-alpine

RUN apk add gettext

COPY ./kickstart /usr/share/nginx/html/kickstart

COPY start.sh /start.sh

CMD [ "/start.sh" ]
