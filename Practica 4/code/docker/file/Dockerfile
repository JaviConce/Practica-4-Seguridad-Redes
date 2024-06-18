FROM midebian 

WORKDIR /app

COPY /file /app/
COPY Cert.crt /app/
COPY Key.key /app/
COPY data/ /app/data/


COPY entrypoints.sh /app/

RUN chmod +x /app/entrypoints.sh

RUN chmod +x /app/data/

ENTRYPOINT [ "/bin/bash", "/app/entrypoints.sh" ]
