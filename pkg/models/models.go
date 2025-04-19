package models

type LogLocalidade struct {
	LocNu       int64   `json:"LOC_NU"`
	UfeSg       string  `json:"UFE_SG"`
	LocNo       string  `json:"LOC_NO"`
	Cep         *string `json:"CEP"`
	LocInSit    string  `json:"LOC_IN_SIT"`
	LocInTipoLoc string  `json:"LOC_IN_TIPO_LOC"`
	LocNuSub    *int64  `json:"LOC_NU_SUB"`
	LocNoAbrev  *string `json:"LOC_NO_ABREV"`
	MunNu       *string `json:"MUN_NU"`
}

type LogBairro struct {
	BaiNu      int64   `json:"BAI_NU"`
	UfeSg      string  `json:"UFE_SG"`
	LocNu      int64   `json:"LOC_NU"`
	BaiNo      string  `json:"BAI_NO"`
	BaiNoAbrev *string `json:"BAI_NO_ABREV"`
}

type LogFaixaUF struct {
	UfeSg     string `json:"UFE_SG"`
	UfeCepIni string `json:"UFE_CEP_INI"`
	UfeCepFim string `json:"UFE_CEP_FIM"`
}

type LogVarLoc struct {
	LocNu int64  `json:"LOC_NU"`
	ValNu int64  `json:"VAL_NU"`
	ValTx string `json:"VAL_TX"`
}

type LogFaixaLocalidade struct {
	LocNu       int64  `json:"LOC_NU"`
	LocCepIni   string `json:"LOC_CEP_INI"`
	LocCepFim   string `json:"LOC_CEP_FIM"`
	LocTipoFaixa string `json:"LOC_TIPO_FAIXA"`
}

type LogVarBai struct {
	BaiNu int64  `json:"BAI_NU"`
	VdbNu int64  `json:"VDB_NU"`
	VdbTx string `json:"VDB_TX"`
}

type LogFaixaBairro struct {
	BaiNu     int64  `json:"BAI_NU"`
	FcbCepIni string `json:"FCB_CEP_INI"`
	FcbCepFim string `json:"FCB_CEP_FIM"`
}

type LogCPC struct {
	CpcNu       int64  `json:"CPC_NU"`
	UfeSg       string `json:"UFE_SG"`
	LocNu       int64  `json:"LOC_NU"`
	CpcNo        string `json:"CPC_NO"`
	CpcEndereco string `json:"CPC_ENDERECO"`
	Cep         string `json:"CEP"`
}

type LogFaixaCPC struct {
	CpcNu      int64  `json:"CPC_NU"`
	CpcInicial string `json:"CPC_INICIAL"`
	CpcFinal   string `json:"CPC_FINAL"`
}

type LogLogradouro struct {
	LogNu        int64   `json:"LOG_NU"`
	UfeSg        string  `json:"UFE_SG"`
	LocNu        int64   `json:"LOC_NU"`
	BaiNuIni     int64   `json:"BAI_NU_INI"`
	BaiNuFim     *int64  `json:"BAI_NU_FIM"`
	LogNo        string  `json:"LOG_NO"`
	LogComplemento *string `json:"LOG_COMPLEMENTO"`
	Cep          string  `json:"CEP"`
	TloTx        string  `json:"TLO_TX"`
	LogStaTlo    *string `json:"LOG_STA_TLO"`
	LogNoAbrev   *string `json:"LOG_NO_ABREV"`
}

type LogVarLog struct {
	LogNu int64  `json:"LOG_NU"`
	VloNu int64  `json:"VLO_NU"`
	TloTx string `json:"TLO_TX"`
	VloTx string `json:"VLO_TX"`
}

type LogNumSec struct {
	LogNu     int64  `json:"LOG_NU"`
	SecNuIni  string `json:"SEC_NU_INI"`
	SecNuFim  string `json:"SEC_NU_FIM"`
	SecInLado string `json:"SEC_IN_LADO"`
}

type LogGrandeUsuario struct {
	GruNu      int64   `json:"GRU_NU"`
	UfeSg      string  `json:"UFE_SG"`
	LocNu      int64   `json:"LOC_NU"`
	BaiNu      int64   `json:"BAI_NU"`
	LogNu      *int64  `json:"LOG_NU"`
	GruNo      string  `json:"GRU_NO"`
	GruEndereco string  `json:"GRU_ENDERECO"`
	Cep        string  `json:"CEP"`
	GruNoAbrev *string `json:"GRU_NO_ABREV"`
}

type LogUnidOper struct {
	UopNu      int64   `json:"UOP_NU"`
	UfeSg      string  `json:"UFE_SG"`
	LocNu      int64   `json:"LOC_NU"`
	BaiNu      int64   `json:"BAI_NU"`
	LogNu      *int64  `json:"LOG_NU"`
	UopNo      string  `json:"UOP_NO"`
	UopEndereco string  `json:"UOP_ENDERECO"`
	Cep        string  `json:"CEP"`
	UopInCp    string  `json:"UOP_IN_CP"`
	UopNoAbrev *string `json:"UOP_NO_ABREV"`
}

type LogFaixaUOP struct {
	UopNu      int64 `json:"UOP_NU"`
	FncInicial int64 `json:"FNC_INICIAL"`
	FncFinal   int64 `json:"FNC_FINAL"`
}

type EctPais struct {
	PaiSg           string `json:"PAI_SG"`
	PaiSgAlternativa string `json:"PAI_SG_ALTERNATIVA"`
	PaiNoPortugues  string `json:"PAI_NO_PORTUGUES"`
	PaiNoIngles     string `json:"PAI_NO_INGLES"`
	PaiNoFrances    string `json:"PAI_NO_FRANCES"`
	PaiAbreviatura  string `json:"PAI_ABREVIATURA"`
}
