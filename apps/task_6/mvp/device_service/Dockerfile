# Сборка и публикация приложения
FROM mcr.microsoft.com/dotnet/sdk:8.0 AS build

WORKDIR /src

COPY device_service.sln ./
COPY device_service/device_service.csproj ./device_service/

RUN dotnet restore device_service.sln

COPY . .

WORKDIR /src/device_service

RUN dotnet publish ./device_service.csproj -c Release -o /app/publish --no-restore

# Итоговый контейнер
FROM mcr.microsoft.com/dotnet/aspnet:8.0 AS runtime

WORKDIR /app

COPY --from=build /app/publish .

ENV ASPNETCORE_URLS=http://+:5002
ENV ASPNETCORE_ENVIRONMENT=Production

EXPOSE 5002

ENTRYPOINT ["dotnet", "device_service.dll"]
