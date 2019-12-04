import axios from "axios";
import { ApiResponse, SocketMessage } from "../types";
import { API_URL } from "../config";

async function xhr<T>(
  route: string,
  params: any,
  verb: string,
  authToken = null
): Promise<ApiResponse<T>> {
  // Getting the Host
  const host = `https://${API_URL}`;
  // Create the URL
  const url = `${host}${route}`;
  console.log(`[${verb}] ${url}`);
  // Createa the option
  const options: any = Object.assign(
    { url, method: verb },
    params ? { data: params } : null
  );
  // setting up the header
  options.headers = {
    Accept: "application/json",
    "Content-Type": "application/json"
  };

  const httpClient = axios.create();
  httpClient.defaults.timeout = 10000;

  return axios
    .request(options)
    .then(response => {
      setTimeout(() => null, 0);
      const resp = { data: response.data } as ApiResponse<T>;
      return resp;
    })
    .catch(error => {
      if (error.response) {
        // The request was made and the server responded with a status
        // code that falls out of the range of 2xx
        if (error.response.data.message) {
          return { errors: error.response.data } as ApiResponse<T>;
        }
        return {
          errors: [
            "Oops! We are experiencing an error. Please try again later."
          ]
        } as ApiResponse<T>;
      }
      if (error.request) {
        // The request was made but no response was received
        // `error.request` is an instance of XMLHttpRequest in the browser and an instance of
        return { errors: [error.message] } as ApiResponse<T>;
      }
      return { errors: [error.message] } as ApiResponse<T>;
    });
}

function stream(params: {
  route: string;
  onData: (results: SocketMessage) => void;
  onComplete: () => void;
}): WebSocket {
  const host = `ws://${API_URL}`;
  const url = `${host}${params.route}`;

  const ws = new WebSocket(url);
  ws.onopen = () => {};

  ws.onmessage = evt => {
    const resp = JSON.parse(evt.data) as SocketMessage;
    params.onData(resp);
  };

  ws.onclose = () => {
    params.onComplete();
  };

  return ws;
}

async function get<T>(
  route: string,
  authToken = null
): Promise<ApiResponse<T>> {
  return xhr<T>(route, null, "GET", authToken);
}

export const ApiService = {
  get,
  stream
};
