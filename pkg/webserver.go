package main

import (
	"fmt"
	"github.com/go-macaron/binding"
	"github.com/go-macaron/cache"
	"github.com/go-macaron/session"
	"gopkg.in/macaron.v1"
	"html/template"
	"net/http"
)

//HTTPConfig has webserver config options
type HTTPConfig struct {
	Port          int    `toml:"port"`
	AdminUser     string `toml:"adminuser"`
	AdminPassword string `toml:"adminpassword"`
}

//UserLogin for login purposes
type UserLogin struct {
	UserName string `form:"username" binding:"Required"`
	Password string `form:"password" binding:"Required"`
}

func webServer(port int) {

	bind := binding.Bind

	/*	jwtMiddleware := jwtmiddleware.New(jwtmiddleware.Options{
		ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
			return []byte("My Secret"), nil
		},
		// When set, the middleware verifies that tokens are signed with the specific signing algorithm
		// If the signing method is not constant the ValidationKeyGetter callback can be used to implement additional checks
		// Important to avoid security issues described here: https://auth0.com/blog/2015/03/31/critical-vulnerabilities-in-json-web-token-libraries/
		SigningMethod: jwt.SigningMethodHS256,
	})*/
	// initiate the app
	m := macaron.Classic()

	// register middleware
	m.Use(macaron.Recovery())
	//	m.Use(gzip.Gziper())
	m.Use(macaron.Static("public",
		macaron.StaticOptions{
			// Prefix is the optional prefix used to serve the static directory content. Default is empty string.
			Prefix: "public",
			// SkipLogging will disable [Static] log messages when a static file is served. Default is false.
			SkipLogging: true,
			// IndexFile defines which file to serve as index if it exists. Default is "index.html".
			IndexFile: "index.html",
			// Expires defines which user-defined function to use for producing a HTTP Expires Header. Default is nil.
			// https://developers.google.com/speed/docs/insights/LeverageBrowserCaching
			Expires: func() string { return "max-age=0" },
		}))
	m.Use(session.Sessioner(session.Options{
		// Name of provider. Default is "memory".
		Provider: "memory",
		// Provider configuration, it's corresponding to provider.
		ProviderConfig: "",
		// Cookie name to save session ID. Default is "MacaronSession".
		CookieName: "MacaronSession",
		// Cookie path to store. Default is "/".
		CookiePath: "/",
		// GC interval time in seconds. Default is 3600.
		Gclifetime: 3600,
		// Max life time in seconds. Default is whatever GC interval time is.
		Maxlifetime: 3600,
		// Use HTTPS only. Default is false.
		Secure: false,
		// Cookie life time. Default is 0.
		CookieLifeTime: 0,
		// Cookie domain name. Default is empty.
		Domain: "",
		// Session ID length. Default is 16.
		IDLength: 16,
		// Configuration section name. Default is "session".
		Section: "session",
	}))

	m.Use(macaron.Renderer(macaron.RenderOptions{
		// Directory to load templates. Default is "templates".
		Directory: "pkg/templates",
		// Extensions to parse template files from. Defaults are [".tmpl", ".html"].
		Extensions: []string{".tmpl", ".html"},
		// Funcs is a slice of FuncMaps to apply to the template upon compilation. Default is [].
		Funcs: []template.FuncMap{map[string]interface{}{
			"AppName": func() string {
				return "snmpcollector"
			},
			"AppVer": func() string {
				return "0.1.0"
			},
		}},
		// Delims sets the action delimiters to the specified strings. Defaults are ["{{", "}}"].
		Delims: macaron.Delims{"{{", "}}"},
		// Appends the given charset to the Content-Type header. Default is "UTF-8".
		Charset: "UTF-8",
		// Outputs human readable JSON. Default is false.
		IndentJSON: true,
		// Outputs human readable XML. Default is false.
		IndentXML: true,
		// Prefixes the JSON output with the given bytes. Default is no prefix.
		// PrefixJSON: []byte("macaron"),
		// Prefixes the XML output with the given bytes. Default is no prefix.
		// PrefixXML: []byte("macaron"),
		// Allows changing of output to XHTML instead of HTML. Default is "text/html".
		HTMLContentType: "text/html",
	}))
	m.Use(cache.Cacher(cache.Options{
		// Name of adapter. Default is "memory".
		Adapter: "memory",
		// Adapter configuration, it's corresponding to adapter.
		AdapterConfig: "",
		// GC interval time in seconds. Default is 60.
		Interval: 60,
		// Configuration section name. Default is "cache".
		Section: "cache",
	}))

	m.Post("/session/create", bind(UserLogin{}), myLoginHandler)

	// Data sources
	m.Get("/runtimeinfo", GetRuntimeInfo)

	m.Group("/snmpdevice", func() {
		m.Get("/", GetSNMPDevices)
		m.Post("/", bind(SnmpDeviceCfg{}), AddSNMPDevice)
		m.Put("/:id", bind(SnmpDeviceCfg{}), UpdateSNMPDevice)
		m.Delete("/:id", DeleteSNMPDevice)
		m.Get("/:id", GetSNMPDeviceByID)
	})

	m.Group("/metric", func() {
		m.Get("/", GetMetrics)
		m.Post("/", bind(SnmpMetricCfg{}), AddMetric)
		m.Put("/:id", bind(SnmpMetricCfg{}), UpdateMetric)
		m.Delete("/:id", DeleteMetric)
		m.Get("/:id", GetMetricByID)
	})

	m.Group("/measurement", func() {
		m.Get("/", GetMeas)
		m.Post("/", bind(InfluxMeasurementCfg{}), AddMeas)
		m.Put("/:id", bind(InfluxMeasurementCfg{}), UpdateMeas)
		m.Delete("/:id", DeleteMeas)
		m.Get("/:id", GetMeasByID)
	})

	m.Group("/measgroups", func() {
		m.Get("/", GetMeasGroup)
		m.Post("/", bind(MGroupsCfg{}), AddMeasGroup)
		m.Put("/:id", bind(MGroupsCfg{}), UpdateMeasGroup)
		m.Delete("/:id", DeleteMeasGroup)
		m.Get("/:id", GetMeasGroupByID)
	})

	m.Group("/measfilters", func() {
		m.Get("/", GetMeasFilter)
		m.Post("/", bind(MeasFilterCfg{}), AddMeasFilter)
		m.Put("/:id", bind(MeasFilterCfg{}), UpdateMeasFilter)
		m.Delete("/:id", DeleteMeasFilter)
		m.Get("/:id", GetMeasFilterByID)
	})

	m.Group("/influxservers", func() {
		m.Get("/", GetInfluxServer)
		m.Post("/", bind(InfluxCfg{}), AddInfluxServer)
		m.Put("/:id", bind(InfluxCfg{}), UpdateInfluxServer)
		m.Delete("/:id", DeleteInfluxServer)
		m.Get("/:id", GetInfluxServerByID)
		m.Get("/ckeckondel/:id", GetInfluxAffectOnDel)
	})

	log.Printf("Server is running on localhost:%d...", port)
	httpServer := fmt.Sprintf("0.0.0.0:%d", port)
	log.Println(http.ListenAndServe(httpServer, m))
}

/****************/
/*Runtime Info
/****************/

func GetRuntimeInfo(ctx *macaron.Context) {
	log.Debugf("Got device runtime info s %+v", &devices)
	ctx.JSON(200, &devices)
}

/****************/
/*SNMP DEVICES
/****************/

// GetSNMPDevices Return snmpdevice list to frontend
func GetSNMPDevices(ctx *macaron.Context) {
	devcfgarray, err := cfg.Database.GetSnmpDeviceCfgArray("")
	if err != nil {
		ctx.JSON(404, err)
		log.Errorf("Error on get Devices :%+s", err)
		return
	}
	ctx.JSON(200, &devcfgarray)
	log.Debugf("Getting DEVICEs %+v", &devcfgarray)
}

// AddSNMPDevice Insert new snmpdevice to de internal BBDD --pending--
func AddSNMPDevice(ctx *macaron.Context, dev SnmpDeviceCfg) {
	log.Printf("ADDING DEVICE %+v", dev)
	affected, err := cfg.Database.AddSnmpDeviceCfg(dev)
	if err != nil {
		log.Warningf("Error on insert for device %s  , affected : %+v , error: %s", dev.ID, affected, err)
		ctx.JSON(404, err)
	} else {
		//TODO: review if needed return data  or affected
		ctx.JSON(200, &dev)
	}
}

// UpdateSNMPDevice --pending--
func UpdateSNMPDevice(ctx *macaron.Context, dev SnmpDeviceCfg) {
	id := ctx.Params(":id")
	log.Debugf("Tying to update: %+v", dev)
	affected, err := cfg.Database.UpdateSnmpDeviceCfg(id, dev)
	if err != nil {
		log.Warningf("Error on update for device %s  , affected : %+v , error: %s", dev.ID, affected, err)
		ctx.JSON(404, err)
	} else {
		//TODO: review if needed return device data
		ctx.JSON(200, &dev)
	}
}

//DeleteSNMPDevice --pending--
func DeleteSNMPDevice(ctx *macaron.Context) {
	id := ctx.Params(":id")
	log.Debugf("Tying to delete: %+v", id)
	affected, err := cfg.Database.DelSnmpDeviceCfg(id)
	if err != nil {
		log.Warningf("Error on delete1 for device %s  , affected : %+v , error: %s", id, affected, err)
		ctx.JSON(404, err)
	} else {
		ctx.JSON(200, "deleted")
	}
}

//GetSNMPDeviceByID --pending--
func GetSNMPDeviceByID(ctx *macaron.Context) {
	id := ctx.Params(":id")
	dev, err := cfg.Database.GetSnmpDeviceCfgByID(id)
	if err != nil {
		log.Warningf("Error on get Device  for device %s  , error: %s", id, err)
		ctx.JSON(404, err)
	} else {
		ctx.JSON(200, &dev)
	}
}

/****************/
/*SNMP METRICS
/****************/

// GetMetrics Return metrics list to frontend
func GetMetrics(ctx *macaron.Context) {
	cfgarray, err := cfg.Database.GetSnmpMetricCfgArray("")
	if err != nil {
		ctx.JSON(404, err)
		log.Errorf("Error on get Metrics :%+s", err)
		return
	}
	ctx.JSON(200, &cfgarray)
	log.Debugf("Getting Metrics %+v", &cfgarray)
}

// AddMetric Insert new metric to de internal BBDD --pending--
func AddMetric(ctx *macaron.Context, dev SnmpMetricCfg) {
	log.Printf("ADDING Metric %+v", dev)
	affected, err := cfg.Database.AddSnmpMetricCfg(dev)
	if err != nil {
		log.Warningf("Error on insert Metric %s  , affected : %+v , error: %s", dev.ID, affected, err)
		ctx.JSON(404, err)
	} else {
		//TODO: review if needed return data  or affected
		ctx.JSON(200, &dev)
	}
}

// UpdateMetric --pending--
func UpdateMetric(ctx *macaron.Context, dev SnmpMetricCfg) {
	id := ctx.Params(":id")
	log.Debugf("Tying to update: %+v", dev)
	affected, err := cfg.Database.UpdateSnmpMetricCfg(id, dev)
	if err != nil {
		log.Warningf("Error on update Metric %s  , affected : %+v , error: %s", dev.ID, affected, err)
		ctx.JSON(404, err)
	} else {
		//TODO: review if needed return device data
		ctx.JSON(200, &dev)
	}
}

//DeleteMetric --pending--
func DeleteMetric(ctx *macaron.Context) {
	id := ctx.Params(":id")
	log.Debugf("Tying to delete: %+v", id)
	affected, err := cfg.Database.DelSnmpMetricCfg(id)
	if err != nil {
		log.Warningf("Error on delete Metric %s  , affected : %+v , error: %s", id, affected, err)
		ctx.JSON(404, err)
	} else {
		ctx.JSON(200, "deleted")
	}
}

//GetMetricByID --pending--
func GetMetricByID(ctx *macaron.Context) {
	id := ctx.Params(":id")
	dev, err := cfg.Database.GetSnmpMetricCfgByID(id)
	if err != nil {
		log.Warningf("Error on get Metric  for device %s  , error: %s", id, err)
		ctx.JSON(404, err)
	} else {
		ctx.JSON(200, &dev)
	}
}

/****************/
/*INFLUX MEASUREMENTS
/****************/

// GetMeas Return measurements list to frontend
func GetMeas(ctx *macaron.Context) {
	cfgarray, err := cfg.Database.GetInfluxMeasurementCfgArray("")
	if err != nil {
		ctx.JSON(404, err)
		log.Errorf("Error on get Influx Measurements :%+s", err)
		return
	}
	ctx.JSON(200, &cfgarray)
	log.Debugf("Getting Measurements %+v", &cfgarray)
}

// AddMeas Insert new measurement to de internal BBDD --pending--
func AddMeas(ctx *macaron.Context, dev InfluxMeasurementCfg) {
	log.Printf("ADDING Measurement %+v", dev)
	affected, err := cfg.Database.AddInfluxMeasurementCfg(dev)
	if err != nil {
		log.Warningf("Error on insert Measurement %s  , affected : %+v , error: %s", dev.ID, affected, err)
		ctx.JSON(404, err)
	} else {
		//TODO: review if needed return data  or affected
		ctx.JSON(200, &dev)
	}
}

// UpdateMeas --pending--
func UpdateMeas(ctx *macaron.Context, dev InfluxMeasurementCfg) {
	id := ctx.Params(":id")
	log.Debugf("Tying to update: %+v", dev)
	affected, err := cfg.Database.UpdateInfluxMeasurementCfg(id, dev)
	if err != nil {
		log.Warningf("Error on update Measurement %s  , affected : %+v , error: %s", dev.ID, affected, err)
		ctx.JSON(404, err)
	} else {
		//TODO: review if needed return device data
		ctx.JSON(200, &dev)
	}
}

//DeleteMeas --pending--
func DeleteMeas(ctx *macaron.Context) {
	id := ctx.Params(":id")
	log.Debugf("Tying to delete: %+v", id)
	affected, err := cfg.Database.DelInfluxMeasurementCfg(id)
	if err != nil {
		log.Warningf("Error on delete Measurement %s  , affected : %+v , error: %s", id, affected, err)
		ctx.JSON(404, err)
	} else {
		ctx.JSON(200, "deleted")
	}
}

//GetMeasByID --pending--
func GetMeasByID(ctx *macaron.Context) {
	id := ctx.Params(":id")
	dev, err := cfg.Database.GetInfluxMeasurementCfgByID(id)
	if err != nil {
		log.Warningf("Error on get Measurement  for device %s  , error: %s", id, err)
		ctx.JSON(404, err)
	} else {
		ctx.JSON(200, &dev)
	}
}

/****************/
/*MEASUREMENT GROUPS
/****************/

// GetMeasGroup Return measurements groups list to frontend
func GetMeasGroup(ctx *macaron.Context) {
	cfgarray, err := cfg.Database.GetMGroupsCfgArray("")
	if err != nil {
		ctx.JSON(404, err)
		log.Errorf("Error on get Measurement Group :%+s", err)
		return
	}
	ctx.JSON(200, &cfgarray)
	log.Debugf("Getting Meas Group %+v", &cfgarray)
}

// AddMeasGroup Insert new measurement groups to de internal BBDD --pending--
func AddMeasGroup(ctx *macaron.Context, dev MGroupsCfg) {
	log.Printf("ADDING Measurement Group %+v", dev)
	affected, err := cfg.Database.AddMGroupsCfg(dev)
	if err != nil {
		log.Warningf("Error on insert Measurement Group %s  , affected : %+v , error: %s", dev.ID, affected, err)
		ctx.JSON(404, err)
	} else {
		//TODO: review if needed return data  or affected
		ctx.JSON(200, &dev)
	}
}

// UpdateMeasGroup --pending--
func UpdateMeasGroup(ctx *macaron.Context, dev MGroupsCfg) {
	id := ctx.Params(":id")
	log.Debugf("Tying to update: %+v", dev)
	affected, err := cfg.Database.UpdateMGroupsCfg(id, dev)
	if err != nil {
		log.Warningf("Error on update Measurement Group %s  , affected : %+v , error: %s", dev.ID, affected, err)
		ctx.JSON(404, err)
	} else {
		//TODO: review if needed return device data
		ctx.JSON(200, &dev)
	}
}

//DeleteMeasGroup --pending--
func DeleteMeasGroup(ctx *macaron.Context) {
	id := ctx.Params(":id")
	log.Debugf("Tying to delete: %+v", id)
	affected, err := cfg.Database.DelMGroupsCfg(id)
	if err != nil {
		log.Warningf("Error on delete Measurement Group %s  , affected : %+v , error: %s", id, affected, err)
		ctx.JSON(404, err)
	} else {
		ctx.JSON(200, "deleted")
	}
}

//GetMeasGroupByID --pending--
func GetMeasGroupByID(ctx *macaron.Context) {
	id := ctx.Params(":id")
	dev, err := cfg.Database.GetMGroupsCfgByID(id)
	if err != nil {
		log.Warningf("Error on get Measurement Group for device %s  , error: %s", id, err)
		ctx.JSON(404, err)
	} else {
		ctx.JSON(200, &dev)
	}
}

/********************/
/*MEASUREMENT FILTERS
/********************/

// GetMeasFilter Return measurements groups list to frontend
func GetMeasFilter(ctx *macaron.Context) {
	cfgarray, err := cfg.Database.GetMeasFilterCfgArray("")
	if err != nil {
		ctx.JSON(404, err)
		log.Errorf("Error on get Measurement Filter :%+s", err)
		return
	}
	ctx.JSON(200, &cfgarray)
	log.Debugf("Getting Measurement Filter %+v", &cfgarray)
}

// AddMeasFilter Insert new measurement groups to de internal BBDD --pending--
func AddMeasFilter(ctx *macaron.Context, dev MeasFilterCfg) {
	log.Printf("ADDING measurement Filter %+v", dev)
	affected, err := cfg.Database.AddMeasFilterCfg(dev)
	if err != nil {
		log.Warningf("Error on insert Measurment Filter %s  , affected : %+v , error: %s", dev.ID, affected, err)
		ctx.JSON(404, err)
	} else {
		//TODO: review if needed return data  or affected
		ctx.JSON(200, &dev)
	}
}

// UpdateMeasFilter --pending--
func UpdateMeasFilter(ctx *macaron.Context, dev MeasFilterCfg) {
	id := ctx.Params(":id")
	log.Debugf("Tying to update: %+v", dev)
	affected, err := cfg.Database.UpdateMeasFilterCfg(id, dev)
	if err != nil {
		log.Warningf("Error on update Measurment Filter %s  , affected : %+v , error: %s", dev.ID, affected, err)
		ctx.JSON(404, err)
	} else {
		//TODO: review if needed return device data
		ctx.JSON(200, &dev)
	}
}

//DeleteMeasFilter --pending--
func DeleteMeasFilter(ctx *macaron.Context) {
	id := ctx.Params(":id")
	log.Debugf("Tying to delete: %+v", id)
	affected, err := cfg.Database.DelMeasFilterCfg(id)
	if err != nil {
		log.Warningf("Error on delete Measurement Filter %s  , affected : %+v , error: %s", id, affected, err)
		ctx.JSON(404, err)
	} else {
		ctx.JSON(200, "deleted")
	}
}

//GetMeasFilterByID --pending--
func GetMeasFilterByID(ctx *macaron.Context) {
	id := ctx.Params(":id")
	dev, err := cfg.Database.GetMeasFilterCfgByID(id)
	if err != nil {
		log.Warningf("Error on get Measurement Filter  for device %s  , error: %s", id, err)
		ctx.JSON(404, err)
	} else {
		ctx.JSON(200, &dev)
	}
}

/****************/
/* INFLUX SERVERS
/****************/

// GetInfluxServer Return Server Array
func GetInfluxServer(ctx *macaron.Context) {
	cfgarray, err := cfg.Database.GetInfluxCfgArray("")
	if err != nil {
		ctx.JSON(404, err)
		log.Errorf("Error on get Influx db :%+s", err)
		return
	}
	ctx.JSON(200, &cfgarray)
	log.Debugf("Getting DEVICEs %+v", &cfgarray)
}

// AddInfluxServer Insert new measurement groups to de internal BBDD --pending--
func AddInfluxServer(ctx *macaron.Context, dev InfluxCfg) {
	log.Printf("ADDING Influx Backend %+v", dev)
	affected, err := cfg.Database.AddInfluxCfg(dev)
	if err != nil {
		log.Warningf("Error on insert new Backend %s  , affected : %+v , error: %s", dev.ID, affected, err)
		ctx.JSON(404, err)
	} else {
		//TODO: review if needed return data  or affected
		ctx.JSON(200, &dev)
	}
}

// UpdateInfluxServer --pending--
func UpdateInfluxServer(ctx *macaron.Context, dev InfluxCfg) {
	id := ctx.Params(":id")
	log.Debugf("Tying to update: %+v", dev)
	affected, err := cfg.Database.UpdateInfluxCfg(id, dev)
	if err != nil {
		log.Warningf("Error on update Influx db %s  , affected : %+v , error: %s", dev.ID, affected, err)
	} else {
		//TODO: review if needed return device data
		ctx.JSON(200, &dev)
	}
}

//DeleteInfluxServer --pending--
func DeleteInfluxServer(ctx *macaron.Context) {
	id := ctx.Params(":id")
	log.Debugf("Tying to delete: %+v", id)
	affected, err := cfg.Database.DelInfluxCfg(id)
	if err != nil {
		log.Warningf("Error on delete influx db %s  , affected : %+v , error: %s", id, affected, err)
		ctx.JSON(404, err)
	} else {
		ctx.JSON(200, "deleted")
	}
}

//GetInfluxServerByID --pending--
func GetInfluxServerByID(ctx *macaron.Context) {
	id := ctx.Params(":id")
	dev, err := cfg.Database.GetInfluxCfgByID(id)
	if err != nil {
		log.Warningf("Error on get Influx db data for device %s  , error: %s", id, err)
		ctx.JSON(404, err)
	} else {
		ctx.JSON(200, &dev)
	}
}

//GetInfluxAffectOnDel --pending--
func GetInfluxAffectOnDel(ctx *macaron.Context) {
	id := ctx.Params(":id")
	obarray, err := cfg.Database.GetInfluxCfgAffectOnDel(id)
	if err != nil {
		log.Warningf("Error on get object array for influx device %s  , error: %s", id, err)
		ctx.JSON(404, err)
	} else {
		ctx.JSON(200, &obarray)
	}
}

/****************/
/*LOGIN
/****************/

func myLoginHandler(ctx *macaron.Context, user UserLogin) {
	fmt.Printf("USER LOGIN: USER: +%#v (Config: %#v)", user, cfg.HTTP)
	if user.UserName == cfg.HTTP.AdminUser && user.Password == cfg.HTTP.AdminPassword {
		fmt.Println("OK")
		ctx.JSON(200, "OK")
	} else {
		fmt.Println("ERROR")
		ctx.JSON(404, "ERROR")
	}
}
