FROM node:alpine
WORKDIR /app
COPY . /app/
RUN npm install && npx tsc
CMD ["node", "./dist/app.js"]