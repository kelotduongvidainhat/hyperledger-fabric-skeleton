FROM node:22-alpine

WORKDIR /app

# Install dependencies
COPY package*.json ./
RUN npm install

# Copy source code
COPY . .

# Expose port
EXPOSE 5173

# Default environment variable for backend visibility from browser
ENV VITE_API_URL=http://localhost:3000

# Start Vite dev server
# Using --host to allow external access within Docker network
CMD ["npm", "run", "dev", "--", "--host"]
