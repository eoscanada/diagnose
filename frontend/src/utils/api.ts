import axios from 'axios';
import { ApiResponse } from '../types'

export const ApiService = {
  get,
  stream,
};

async function get<T>(route:string, authToken = null): Promise<ApiResponse<T>> {
  return xhr<T>(route, null, 'GET', authToken);
}

async function xhr<T>(route:string, params:any, verb:string, authToken = null): Promise<ApiResponse<T>> {
  // Getting the Host
  const host = 'http://localhost:8080/api/'
  // Create the URL
  const url = `${host}${route}`
  console.log(`[${verb}] ${url}`)
  // Createa the option
  let options:any = Object.assign({ url: url, method: verb }, params ? { data: params } : null );
  // setting up the header
  options.headers = {
      'Accept': 'application/json',
      'Content-Type': 'application/json',
    };

  const httpClient = axios.create();
  httpClient.defaults.timeout = 10000;

  return axios.request(options)
  .then(function (response) {
    setTimeout(() => null, 0);
    let resp = {data: response.data} as ApiResponse<T>
    console.log("resp: ", resp)
    return resp
  })
  .catch(function (error) {
    if (error.response) {
      // The request was made and the server responded with a status
      // code that falls out of the range of 2xx
      if(error.response.data.message){
        return { errors: error.response.data } as ApiResponse<T>
      }else{
        return { errors: ["Oops! We are experiencing an error. Please try again later."] } as ApiResponse<T>
      }
    } else if (error.request) {
      // The request was made but no response was received
      // `error.request` is an instance of XMLHttpRequest in the browser and an instance of
      return { errors: [error.message] } as ApiResponse<T>
    } else {
      // Something happened in setting up the request that triggered an Error
      return { errors: [error.message] } as ApiResponse<T>
    }
    throw new Error('unsupported Error Reponse')
  });
}


function stream<T>(params: {
  route: string
  onData: (results: ApiResponse<T>) => void
  onComplete: () => void
}):WebSocket {
  const host = 'ws://localhost:8080/'
  const url = `${host}${params.route}`

  let ws = new WebSocket(url);
  ws.onopen = () => {}

  ws.onmessage = evt => {
    let resp = {data: JSON.parse(evt.data)} as ApiResponse<T>
    params.onData(resp)
  };

  ws.onclose = () => {
    params.onComplete()
  }

  return ws
}



