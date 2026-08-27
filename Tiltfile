load('ext://nerdctl', 'nerdctl_build')

ORG_NAME = 'iamnande'
SERVICE_NAME = 'pings'
PROJECT_REPO_URL = 'github.com/iamnande/hyrule'
GO_VERSION_PACKAGE = PROJECT_REPO_URL + '/internal/lib/version'

PROJECT_VERSION = str(read_file('VERSION')).strip()
PROJECT_COMMIT = str(local('git rev-parse HEAD | cut -c1-8', quiet=True)).strip()
BUILD_DATETIME = str(local('date -u +%Y-%m-%dT%H:%M:%SZ', quiet=True)).strip()

nerdctl_build(
    ref='%s/%s' % (ORG_NAME, SERVICE_NAME),
    context='.',
    dockerfile='cmd/Dockerfile',
    build_args={
        'ORG_NAME': ORG_NAME,
        'SERVICE_NAME': SERVICE_NAME,
        'PROJECT_COMMIT': PROJECT_COMMIT,
        'PROJECT_VERSION': PROJECT_VERSION,
        'PROJECT_REPO_URL': PROJECT_REPO_URL,
        'GO_VERSION_PACKAGE': GO_VERSION_PACKAGE,
        'BUILD_DATETIME': BUILD_DATETIME,
    },
)

init_sql = read_file('stack/postgres/init/01-app-role.sql')
init_sql_indented = '\n'.join(['    ' + line for line in str(init_sql).splitlines()])

k8s_yaml(blob("""
apiVersion: v1
kind: ConfigMap
metadata:
  name: pings-postgres-init
data:
  01-app-role.sql: |
%s
""" % init_sql_indented))

k8s_yaml('deploy/local/postgres.yaml')
k8s_resource('pings-postgres', port_forwards='5432:5432', labels=['data'])

local_resource(
    'pings-migrate',
    cmd='make db-migrate-up',
    resource_deps=['pings-postgres'],
    labels=['data'],
)

k8s_yaml(helm('deploy/helm/app-platform', name='pings', values=['deploy/values/pings/values.yaml']))
k8s_resource('pings', resource_deps=['pings-migrate'], port_forwards='8000:8000', labels=['pings'])
