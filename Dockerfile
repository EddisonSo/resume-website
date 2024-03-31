FROM nginx
COPY nginx/nginx.conf /etc/nginx/conf.d/default.conf
COPY index.html /etc/nginx/html/index.html
