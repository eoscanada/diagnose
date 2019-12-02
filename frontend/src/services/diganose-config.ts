import { ApiService } from '../utils/api'
import {ApiResponse, DiagnoseConfig} from '../types'

export const diagnoseService = {
  config,

};

async function config(): Promise<ApiResponse<DiagnoseConfig>> {
  return ApiService.get<DiagnoseConfig>("config")
}