load('ext://nerdctl', 'nerdctl_build')
load('ext://helm_resource', 'helm_resource')

ORG_NAME = 'iamnande'
PROJECT_REPO_URL = 'github.com/iamnande/hyrule'
GO_VERSION_PACKAGE = PROJECT_REPO_URL + '/internal/lib/version'

PROJECT_VERSION = str(read_file('VERSION')).strip()
PROJECT_COMMIT = str(local('git rev-parse HEAD | cut -c1-8', quiet=True)).strip()
BUILD_DATETIME = str(local('date -u +%Y-%m-%dT%H:%M:%SZ', quiet=True)).strip()

service_slugs = [s for s in str(local('ls deploy/values', quiet=True)).strip().split('\n') if s]

def configmap_from_dir(name, path):
    files = [f for f in str(local('ls %s' % path, quiet=True)).strip().split('\n') if f]
    entries = []
    for f in files:
        content = str(read_file('%s/%s' % (path, f))).splitlines()
        indented = '\n'.join(['    ' + line for line in content])
        entries.append('  %s: |\n%s' % (f, indented))
    return blob("""
apiVersion: v1
kind: ConfigMap
metadata:
  name: %s
data:
%s
""" % (name, '\n'.join(entries)))

k8s_yaml(configmap_from_dir('hyrule-database-init', 'stack/postgres/init'))
k8s_yaml('deploy/local/postgres.yaml')
k8s_resource(
    'hyrule-database',
    objects=['hyrule-database-init:configmap'],
    port_forwards='5432:5432',
    labels=['local-only'],
)

k8s_yaml(configmap_from_dir('hyrule-migrations', 'migrations'))
k8s_yaml('deploy/local/migrate-job.yaml')
k8s_resource(
    'migrate',
    objects=['hyrule-migrations:configmap'],
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
