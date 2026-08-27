load('ext://nerdctl', 'nerdctl_build')

ORG_NAME = 'iamnande'
PROJECT_REPO_URL = 'github.com/iamnande/hyrule'
GO_VERSION_PACKAGE = PROJECT_REPO_URL + '/internal/lib/version'

PROJECT_VERSION = str(read_file('VERSION')).strip()
PROJECT_COMMIT = str(local('git rev-parse HEAD | cut -c1-8', quiet=True)).strip()
BUILD_DATETIME = str(local('date -u +%Y-%m-%dT%H:%M:%SZ', quiet=True)).strip()

init_sql = read_file('stack/postgres/init/01-app-role.sql')
init_sql_indented = '\n'.join(['    ' + line for line in str(init_sql).splitlines()])

k8s_yaml(blob("""
apiVersion: v1
kind: ConfigMap
metadata:
  name: hyrule-database-init
data:
  01-app-role.sql: |
%s
""" % init_sql_indented))

k8s_yaml('deploy/local/postgres.yaml')
k8s_resource('hyrule-database', port_forwards='5432:5432', labels=['data'])

local_resource(
    'migrate',
    cmd='make db-migrate-up',
    resource_deps=['hyrule-database'],
    labels=['data'],
)

service_slugs = [s for s in str(local('ls deploy/values', quiet=True)).strip().split('\n') if s]

for slug in service_slugs:
    values_path = 'deploy/values/%s/values.yaml' % slug
    values = read_yaml(values_path)
    app_env = values.get('app', {}).get('env', {})
    needs_db = 'HYRULE_DATABASE_HOST' in app_env

    nerdctl_build(
        ref='%s/%s' % (ORG_NAME, slug),
        context='.',
        dockerfile='cmd/Dockerfile',
        build_args={
            'ORG_NAME': ORG_NAME,
            'SERVICE_NAME': slug,
            'PROJECT_COMMIT': PROJECT_COMMIT,
            'PROJECT_VERSION': PROJECT_VERSION,
            'PROJECT_REPO_URL': PROJECT_REPO_URL,
            'GO_VERSION_PACKAGE': GO_VERSION_PACKAGE,
            'BUILD_DATETIME': BUILD_DATETIME,
        },
    )

    k8s_yaml(helm('deploy/helm/app-platform', name=slug, values=[values_path]))
    k8s_resource(
        slug,
        resource_deps=['migrate'] if needs_db else [],
        port_forwards=port_forward(local_port=0, container_port=8000),
        labels=[slug],
    )
