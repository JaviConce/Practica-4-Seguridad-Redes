FROM midebian 

WORKDIR /app

COPY /auth /app/
COPY Cert.crt /app/
COPY Key.key /app/
COPY user.json /app/


COPY entrypoints.sh /app/

RUN chmod +x /app/entrypoints.sh

RUN chmod +x /app/user.json

ENTRYPOINT [ "/bin/bash", "/app/entrypoints.sh" ]