FROM node:latest as builder

WORKDIR /app
COPY package.json package-lock.json /app/
RUN cd /app && npm set progress=false && npm install
COPY . /app

RUN cd /app && npm run build

FROM nginx:alpine
RUN rm -rf /usr/share/nginx/html/*
# todo(damienr74), put this in configmap.
COPY nginx.conf /etc/nginx/nginx.conf
COPY --from=builder /app/dist/kustomize-search/ /usr/share/nginx/html
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
