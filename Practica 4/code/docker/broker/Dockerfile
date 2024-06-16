FROM midebian 

WORKDIR /app

COPY /broker /app/
COPY Cert.crt /app/
COPY Key.key /app/


COPY entrypoints.sh /app/

RUN chmod +x /app/entrypoints.sh

ENTRYPOINT [ "/bin/bash", "/app/entrypoints.sh" ]