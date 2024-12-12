audience-service
    repo {sholud not delete data but mark as deleted}
        save auditories and integrations to local DB
        get auditories and integrations from local DB
        get applications(requests) form external MySQL DB
    filter
         applications on filters from querry
    controller {stupid simple}
    bl-layer
        for audiences where data range of applications exists it should be unchangable
        for audience where data range of applications do not exist it should be updated BEFORE sending UpdateMsg for rabbit 

    
integration-service
    ~get knowlege about ads-cabinets api requirments
    ~consume from queue
    repo
        log repo update
    bl-layer
        push messages to ads api

reporting-service
    repo
        get data from external DB
    filter
        filter applications by fields{
            status
            duration
            project_id
            project_name
            estate_type
            audience
            city_id
            region
            date_range (сроки)
            expired
        }
        filter managers by fields{
            data (сроки)
            ФИО
        }
    controller {stupid simple api}
    bl-layer
        excel convertation
        create table of report

crm-integration-service
    repo
        get applications from external DB
    bl-layer
        push task to macro api

logging-service
    repo
        get\set logs db
    controller
        get logs
        set log
    bl-layer:
        collect and store logs