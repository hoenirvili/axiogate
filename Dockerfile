FROM scratch
COPY axiogate /axiogate
EXPOSE 8080
CMD ["./axiogate"]
