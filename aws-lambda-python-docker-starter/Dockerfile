FROM public.ecr.aws/lambda/python:3.12

# 1. Install dependencies only (cached unless pyproject.toml changes)
COPY pyproject.toml ${LAMBDA_TASK_ROOT}/
RUN python -m pip install --upgrade pip \
    && python -c "import tomllib, pathlib; f=open('pyproject.toml','rb'); deps=tomllib.load(f)['project']['dependencies']; pathlib.Path('requirements.txt').write_text('\n'.join(deps))" \
    && python -m pip install --no-cache-dir -r requirements.txt \
    && rm -f pyproject.toml requirements.txt

# 2. Copy source code (fast rebuild when only source changes)
COPY app ${LAMBDA_TASK_ROOT}/app

CMD ["app.entrypoints.lambda_handler.handler"]
