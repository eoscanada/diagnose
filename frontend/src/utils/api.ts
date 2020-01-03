import axios from "axios"
import { ApiResponse, SocketMessage, DataApiResponse } from "../types"
import { API_URL, API_SSL } from "../config"

async function xhr<T>(
  route: string,
  params: any,
  verb: string,
  authToken = null
): Promise<ApiResponse<T>> {
  // Getting the Host
  const httpProto = API_SSL ? "https" : "http"
  const host = `${httpProto}://${API_URL}`
  // Create the URL
  const url = `${host}${route}`
  console.log(`[${verb}] ${url}`)
  // Createa the option
  const options: any = Object.assign({ url, method: verb }, params ? { data: params } : null)
  // setting up the header
  options.headers = {
    Accept: "application/json",
    "Content-Type": "application/json"
  }

  const httpClient = axios.create()
  httpClient.defaults.timeout = 10000

  return axios
    .request(options)
    .then<DataApiResponse<T>>((response) => {
      setTimeout(() => null, 0)

      return { type: "data", data: response.data }
    })
    .catch((error) => {
      if (error.response) {
        // The request was made and the server responded with a status
        // code that falls out of the range of 2xx
        if (error.response.data.message) {
          return { type: "error", errors: [error.response.data] }
        }

        return {
          type: "error",
          errors: ["Oops! We are experiencing an error. Please try again later."]
        }
      }

      if (error.request) {
        // The request was made but no response was received
        // `error.request` is an instance of XMLHttpRequest in the browser and an instance of
        return { type: "error", errors: [error.message] }
      }

      return { type: "error", errors: [error] }
    })
}

function stream(params: {
  route: string
  onData: (results: SocketMessage) => void
  onComplete: () => void
  onError?: (error: Event) => void
}): WebSocket {
  const wsProto = API_SSL ? "wss" : "ws"
  const host = `${wsProto}://${API_URL}`
  const url = `${host}${params.route}`

  const ws = new WebSocket(url)
  ws.onopen = () => {}

  ws.onerror = (error) => {
    if (params.onError) {
      params.onError(error)
    }
  }

  ws.onmessage = (evt) => {
    const resp = JSON.parse(evt.data) as SocketMessage
    params.onData(resp)
  }

  ws.onclose = () => {
    params.onComplete()
  }

  return ws
}

async function get<T>(route: string, authToken = null): Promise<ApiResponse<T>> {
  return xhr<T>(route, null, "GET", authToken)
}

export const ApiService = {
  get,
  stream
}
