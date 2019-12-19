export const API_URL =
  process.env.NODE_ENV === "development"
    ? "localhost:8080/api/"
    : `${window.location.host}/api/`;

export const API_SSL =
  process.env.NODE_ENV === "development"
    ? false
    : window.location.protocol.match(/^https:/) != null;

console.log("Api: ", API_URL, "SSL: ", API_SSL, "Location: ", window.location);
