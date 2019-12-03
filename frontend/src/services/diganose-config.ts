import { ApiService } from "../utils/api";
import { ApiResponse, DiagnoseConfig } from "../types";

async function config(): Promise<ApiResponse<DiagnoseConfig>> {
  return ApiService.get<DiagnoseConfig>("config");
}

export const diagnoseService = {
  config
};
