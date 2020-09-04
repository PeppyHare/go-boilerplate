import {API_URL} from "src/constants";
import {AuthToken} from "src/types";
import HttpError, {AccessDeniedHttpError, NotFoundHttpError, UnauthorizedHttpError,} from "src/errors";

export const fetchJSON = (authToken?: AuthToken) => async (
  path: string,
  method: string,
  params?: { [key: string]: string | Array<string> } | null,
  body?: string
): Promise<any> => {
  const url = new URL(decodeURIComponent(path), API_URL);

  if (params) {
    Object.keys(params).forEach((k) => {
      if (Array.isArray(params[k])) {
        (params[k] as Array<string>).forEach((v) => {
          url.searchParams.set(k, v);
        });
      } else {
        url.searchParams.set(k, params[k] as string);
      }
    });
  }

  const headers: HeadersInit = {
    "Content-Type": "application/json",
  };

  if (authToken) {
    headers["Authorization"] = `Bearer ${authToken}`;
  }

  const response = await fetch(url.toString(), {method, headers, body});

  if (!response.ok) {
    switch (response.status) {
      case 401:
        throw new UnauthorizedHttpError(response.statusText);
      case 403:
        throw new AccessDeniedHttpError(response.statusText);
      case 404:
        throw new NotFoundHttpError(response.statusText);
      default:
        throw new HttpError(response.statusText);
    }
  }

  return response.json();
};
