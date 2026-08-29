load('ext://nerdctl', 'nerdctl_build')
load('ext://helm_resource', 'helm_resource')

ORG_NAME = 'iamnande'
PROJECT_REPO_URL = 'github.com/iamnande/hyrule'
GO_VERSION_PACKAGE = PROJECT_REPO_URL + '/internal/lib/version'

PROJECT_VERSION = str(read_file('VERSION')).strip()
PROJECT_COMMIT = str(local('git rev-parse HEAD | cut -c1-8', quiet=True)).strip()
BUILD_DATETIME = str(local('date -u +%Y-%m-%dT%H:%M:%SZ', quiet=True)).strip()

service_slugs = [s for s in str(local('ls deploy/values', quiet=True)).strip().split('\n') if s]

init_dir = 'stack/postgres/init'
init_files = [f for f in str(local('ls %s' % init_dir, quiet=True)).strip().split('\n') if f]
init_data = []
for init_file in init_files:
    content = str(read_file('%s/%s' % (init_dir, init_file))).splitlines()
    indented = '\n'.join(['    ' + line for line in content])
    init_data.append('  %s: |\n%s' % (init_file, indented))

k8s_yaml(blob("""
apiVersion: v1
kind: ConfigMap
metadata:
  name: hyrule-database-init
data:
%s
""" % '\n'.join(init_data)))

k8s_yaml('deploy/local/postgres.yaml')
k8s_resource(
    'hyrule-database',
    objects=['hyrule-database-init:configmap'],
    port_forwards='5432:5432',
    labels=['local-only'],
)

local_resource(
    'migrate',
    cmd='make db-migrate-up',
    resource_deps=['hyrule-database'],
    labels=['local-only'],
)

for slug in service_slugs:
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

for slug in service_slugs:
    values_path = 'deploy/values/%s/values.yaml' % slug
    values = read_yaml(values_path)
    needs_db = 'HYRULE_DATABASE_HOST' in values.get('app', {}).get('env', {})

    helm_resource(
        name=slug,
        chart='deploy/helm/app-platform',
        release_name=slug,
        flags=['--values=%s' % values_path],
        image_deps=['%s/%s' % (ORG_NAME, slug)],
        image_keys=[('app.image.repository', 'app.image.tag')],
        resource_deps=['migrate'] if needs_db else [],
        port_forwards=[port_forward(local_port=0, container_port=8000)],
        labels=['deploy'],
    )
    k8s_resource(slug, trigger_mode=TRIGGER_MODE_MANUAL)
