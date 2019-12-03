export const API_URL =
  !process.env.NODE_ENV || process.env.NODE_ENV === "development"
    ? "localhost:8080/api/"
    : `${window.location.hostname}/api/`;
console.log("Api: ", API_URL);
