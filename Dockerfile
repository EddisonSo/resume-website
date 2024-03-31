FROM python:3.12.0b4-slim-bullseye

RUN pip install --upgrade pip

COPY ./flaskr opt/flaskr
WORKDIR /opt/flaskr
RUN pip3 install -r requirements.txt
CMD ["gunicorn", "-b", "0.0.0.0", "-w", "4", "--threads", "4", "--access-logfile", "-", "main:app"]
EXPOSE 8000 
