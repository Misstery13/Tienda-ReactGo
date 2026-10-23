# ===== Etapa de compilación del frontend =====
FROM node:22-alpine AS build

WORKDIR /app

COPY package.json package-lock.json ./
RUN npm ci

COPY . .

# Vite congela las variables VITE_* dentro del bundle al compilar.
# Vacío = rutas relativas (/api/...), que es lo que queremos detrás del proxy.
ARG VITE_API_URL=""
ENV VITE_API_URL=$VITE_API_URL
RUN npm run build

# ===== Servidor estático + proxy inverso =====
FROM nginx:alpine

COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=build /app/dist /usr/share/nginx/html

EXPOSE 80
