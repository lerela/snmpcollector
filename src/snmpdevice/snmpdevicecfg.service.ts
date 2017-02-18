import { HttpAPI } from '../common/httpAPI'
import { Injectable } from '@angular/core';
import { Observable } from 'rxjs/Observable';

declare var _:any;

@Injectable()
export class SnmpDeviceService {

    constructor(public httpAPI: HttpAPI) {
        console.log('Task Service created.', httpAPI);
    }

    addDevice(dev) {
        return this.httpAPI.post('/api/cfg/snmpdevice',JSON.stringify(dev,function (key,value) {
            console.log("KEY: "+key+" Value: "+value);
            if ( key == 'Port' ||
            key == 'Retries' ||
            key == 'Timeout' ||
            key == 'Repeat' ||
            key == 'Freq'  ||
            key == 'UpdateFltFreq' ) {
                return parseInt(value);
            }

            if ( key == 'Active' ||
            key == 'SnmpDebug' ||
            key == 'DisableBulk' ) return ( value === "true" || value === true);
            if ( key == 'ExtraTags' ) return  String(value).split(',');
            if ( key == 'MeasFilters' ||
            key == 'MeasurementGroups') {
                if (value == "") return null;
                else return value;
            }
            return value;
        }))
        .map( (responseData) => responseData.json());
    }

    editDevice(dev, id) {
        console.log("DEV: ",dev);
        //TODO: Se tiene que coger el oldid para substituir en la configuración lo que toque!!!!
        return this.httpAPI.put('/api/cfg/snmpdevice/'+id,JSON.stringify(dev,function (key,value) {
            if ( key == 'Port' ||
            key == 'Retries' ||
            key == 'Timeout' ||
            key == 'Repeat' ||
            key == 'Freq'  ||
            key == 'UpdateFltFreq') {
                return parseInt(value);
            }
            if ( key == 'Active' ||
            key == 'SnmpDebug' ||
            key == 'DisableBulk' ) return ( value === "true" || value === true);
            if ( key == 'ExtraTags' ) return  String(value).split(',');
            if ( key == 'MeasFilters' ||
            key == 'MeasurementGroups') {
                if (value == "") return null;
                else return value;
            }
            return value;
        }))
        .map( (responseData) => responseData.json());
    }

    getDevices(filter_s: string) {
        // return an observable
        return this.httpAPI.get('/api/cfg/snmpdevice')
        .map( (responseData) => {
            return responseData.json();
        })
        .map((snmpdevs) => {
            console.log("MAP SERVICE",snmpdevs);
            let result = [];
            if (snmpdevs) {
                _.forEach(snmpdevs,function(value,key){
                    console.log("FOREACH LOOP",value,value.ID);
                    if(filter_s && filter_s.length > 0 ) {
                        console.log("maching: "+value.ID+ "filter: "+filter_s);
                        var re = new RegExp(filter_s, 'gi');
                        if (value.ID.match(re)){
                            result.push(value);
                        }
                        console.log(value.ID.match(re));
                    } else {
                        result.push(value);
                    }
                });
            }
            return result;
        });
    }
    getDevicesById(id : string) {
        // return an observable
        console.log("ID: ",id);
        return this.httpAPI.get('/api/cfg/snmpdevice/'+id)
        .map( (responseData) =>
            responseData.json()
    )};

    checkOnDeleteSNMPDevice(id : string){
    return this.httpAPI.get('/api/cfg/snmpdevice/checkondel/'+id)
    .map( (responseData) =>
     responseData.json()
    ).map((deleteobject) => {
        console.log("MAP SERVICE",deleteobject);
        let result : any = {'ID' : id};
        _.forEach(deleteobject,function(value,key){
            result[value.TypeDesc] = [];
        });
        _.forEach(deleteobject,function(value,key){
            result[value.TypeDesc].Description=value.Action;
            result[value.TypeDesc].push(value.ObID);
        });
        return result;
    });
  };

    deleteDevice(id : string) {
        // return an observable
        console.log("ID: ",id);
        console.log("DELETING");
        return this.httpAPI.delete('/api/cfg/snmpdevice/'+id)
        .map( (responseData) =>
         responseData.json()
        );
    };

    pingDevice(dev) {
        console.log(dev);
        return this.httpAPI.post('/api/rt/agent/snmpconsole/ping/',JSON.stringify(dev,function (key,value) {
            if ( key == 'Port' ||
            key == 'Retries' ||
            key == 'Timeout' ||
            key == 'Repeat' ||
            key == 'Freq' ) {
                return parseInt(value);
            }
            if ( key == 'Active' ||
            key == 'SnmpDebug' ||
            key == 'DisableBulk' ) return ( value === "true" || value === true);
            if ( key == 'ExtraTags' ) return  String(value).split(',');
            if ( key == 'MeasFilters' ||
            key == 'MeasurementGroups') {
                if (value == "") return null;
                else return value;
            }
            return value;
        }))
        .map( (responseData) => responseData.json());
    }

    sendQuery(dev,getMode,oid) {
        return this.httpAPI.post('/api/rt/agent/snmpconsole/query/'+getMode+'/oid/'+oid,JSON.stringify(dev,function (key,value) {
            if ( key == 'Port' ||
            key == 'Retries' ||
            key == 'Timeout' ||
            key == 'Repeat' ||
            key == 'Freq' ) {
                return parseInt(value);
            }
            if ( key == 'Active' ||
            key == 'SnmpDebug' ||
            key == 'DisableBulk' ) return ( value === "true" || value === true);
            if ( key == 'ExtraTags' ) return  String(value).split(',');
            if ( key == 'MeasFilters' ||
            key == 'MeasurementGroups') {
                if (value == "") return null;
                else return value;
            }
            return value;
        }))
        .map( (responseData) => responseData.json());
    }
}
