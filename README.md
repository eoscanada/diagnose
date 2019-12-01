Diagnose
--------

Small tool to diagnose the in-cluster components, quick internal debug
tools.


SetupRoutes
blockHoles
dbHoles
searchHoles


func (e *ETHDiagnose) SetupRoutes(s *mux.Router)
func (e *ETHDiagnose) blockHoles(w http.ResponseWriter, r *http.Request)
func (e *ETHDiagnose) dbHoles(w http.ResponseWriter, r *http.Request)
func (e *ETHDiagnose) searchHoles(w http.ResponseWriter, r *http.Request


func (e *EOSDiagnose) SetupRoutes(s *mux.Router)
func (e *EOSDiagnose) blockHoles(w http.ResponseWriter, r *http.Request)
func (e *EOSDiagnose) dbHoles(w http.ResponseWriter, r *http.Request)
func (e *EOSDiagnose) searchHoles(w http.ResponseWriter, r *http.Request)