version_settings(constraint='>=0.37.0')

load('ext://nerdctl', 'nerdctl_build')
load('ext://helm_resource', 'helm_resource')

ORG_NAME = 'iamnande'
PROJECT_REPO_URL = 'github.com/iamnande/hyrule'
GO_VERSION_PACKAGE = PROJECT_REPO_URL + '/go/internal/lib/version'

PROJECT_VERSION = str(read_file('VERSION')).strip()
PROJECT_COMMIT = str(local('git rev-parse HEAD | cut -c1-8', quiet=True)).strip()
BUILD_DATETIME = str(local('date -u +%Y-%m-%dT%H:%M:%SZ', quiet=True)).strip()

SERVICE_SLUGS = [s for s in str(local('ls deploy/values', quiet=True)).strip().split('\n') if s]


def configmap_from_dir(name, path, namespace):
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
  namespace: %s
data:
%s
""" % (name, namespace, '\n'.join(entries)))


def local_database():
    k8s_yaml(blob("""
apiVersion: v1
kind: Namespace
metadata:
  name: databases
"""))
    k8s_yaml(configmap_from_dir('postgres-init', 'local/postgres/init', namespace='databases'))
    k8s_yaml('local/cluster/postgres.yaml')
    k8s_resource(
        'postgres',
        objects=['postgres-init:configmap', 'databases:namespace'],
        port_forwards='5432:5432',
        labels=['local-only'],
    )


def has_postgres_datastore(datastores):
    return any([
        d.get('flavor') == 'postgres' and d.get('schema_management', 'enabled') != 'disabled'
        for d in datastores
    ])


def local_migrate(slug):
    values_path = 'deploy/values/%s/values.yaml' % slug
    values = read_yaml(values_path)
    datastores = values.get('app', {}).get('datastores', [])
    if not has_postgres_datastore(datastores):
        return
    migrations_configmap = '%s-migrations' % slug  # matches app.fullname-derived naming
    k8s_yaml(blob("""
apiVersion: v1
kind: Namespace
metadata:
  name: %s
""" % slug))
    k8s_yaml(configmap_from_dir(
        migrations_configmap, 'migrations/%s' % slug, namespace=slug,
    ))
    rendered = local(
        ('helm template %s deploy/helm/app-platform --namespace=%s --values=%s ' +
         '--show-only charts/app/templates/migrate.yaml') % (slug, slug, values_path),
        quiet=True,
    )
    k8s_yaml(blob(str(rendered)))
    k8s_resource(
        '%s-migrate' % slug,
        objects=['%s:configmap' % migrations_configmap, '%s:namespace' % slug],
        resource_deps=['postgres'],
        labels=['local-only'],
    )


def build_service(slug):
    nerdctl_build(
        ref='%s/%s' % (ORG_NAME, slug),
        context='.',
        dockerfile='go/cmd/Dockerfile',
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


def deploy_service(slug):
    values_path = 'deploy/values/%s/values.yaml' % slug
    values = read_yaml(values_path)
    needs_migrate = has_postgres_datastore(values.get('app', {}).get('datastores', []))

    helm_resource(
        name=slug,
        chart='deploy/helm/app-platform',
        release_name=slug,
        namespace=slug,
        flags=['--values=%s' % values_path, '--create-namespace', '--set-json=app.datastores=[]'],
        image_deps=['%s/%s' % (ORG_NAME, slug)],
        image_keys=[('app.image.repository', 'app.image.tag')],
        resource_deps=['%s-migrate' % slug] if needs_migrate else [],
        port_forwards=[port_forward(local_port=0, container_port=8000)],
        labels=['deploy'],
    )
    k8s_resource(
        slug,
        trigger_mode=TRIGGER_MODE_MANUAL,
        links=[
            link('/discovery', 'discovery'),
            link('/healthz', 'healthz'),
        ],
    )


print('hyrule: %d service(s) under deploy/values/: %s' % (len(SERVICE_SLUGS), ', '.join(SERVICE_SLUGS)))

local_database()
for slug in SERVICE_SLUGS:
    local_migrate(slug)

for slug in SERVICE_SLUGS:
    build_service(slug)

for slug in SERVICE_SLUGS:
    deploy_service(slug)
