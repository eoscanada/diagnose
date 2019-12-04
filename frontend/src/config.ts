export const API_URL =
  !process.env.NODE_ENV || process.env.NODE_ENV === "development"
    ? "localhost:8080/api/"
    : `${window.location.hostname}/api/`;
export const API_SSL = !process.env.NODE_ENV || process.env.NODE_ENV === "development"
    ? false
    : true;

console.log("Api: ", API_URL, "SSL: ", API_SSL);
