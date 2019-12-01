import { useState, useEffect } from 'react';
import { diagnoseService } from '../services/diganose-config'
import { DiagnoseConfig } from '../types'

export function useAppConfig(): (DiagnoseConfig | undefined) {
  const [config, setConfig] = useState<DiagnoseConfig>();

  useEffect(() => {
    (async () => {
      diagnoseService.config()
        .then(response => {
          setConfig(response.data)
        });
    })();
    return () => {
    }
  },[]);

  return config
}