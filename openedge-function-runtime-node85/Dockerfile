FROM node:8.5.0-alpine 

WORKDIR /home/openedge/bin
COPY . .
RUN chmod +x openedge_function_runtime_node85.js
RUN npm install
WORKDIR /home/openedge
ENTRYPOINT ["bin/openedge_function_runtime_node85.js"]
CMD []