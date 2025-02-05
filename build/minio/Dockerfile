FROM minio/mc

COPY ./init.sh .
RUN ls
RUN chmod +x /init.sh

ENTRYPOINT ["./init.sh"]