"use strict";

const byID = (id) => document.getElementById(id);
const state = { tasks: null, filter: "all", busy: false, loading: false, dialog: null };
const headings = { all: "Все задачи", active: "В работе", done: "Выполнено" };
const createForm = byID("create-form");
const newTitle = byID("new-title");
const taskList = byID("task-list");
const dialog = byID("task-dialog");
const editTitle = byID("edit-title");
const errorBox = byID("error");
const dialogError = byID("dialog-error");
const refresh = byID("refresh");
const filters = [...document.querySelectorAll("[data-filter]")];
const logout = byID("logout");
const accountView = byID("account-view");
const workspaceView = byID("workspace-view");
const sessionCheck = byID("session-check");
const sessionRetry = byID("session-retry");
const authView = byID("auth-view");
const authForm = byID("auth-form");
const authEmail = byID("auth-email");
const authPassword = byID("auth-password");
const authError = byID("auth-error");
const authNotice = byID("auth-notice");
const authSubmit = byID("auth-submit");
const passwordVisibility = byID("password-visibility");
const authModes = [...document.querySelectorAll("[data-auth-mode]")];
let authMode = "login";
let authBusy = false;

const today = new Intl.DateTimeFormat("ru-RU", {
  day: "numeric", month: "long", weekday: "long",
}).format(new Date());
byID("today").textContent = today.charAt(0).toUpperCase() + today.slice(1);

function showError(box, message = "") {
  box.textContent = message;
  box.hidden = message === "";
}

function statusMessage(status) {
  if (status === 404) return "Задача больше не найдена. Обнови список.";
  if (status === 413) return "Название слишком длинное. Попробуй сократить его.";
  if (status === 400) return "Сервер не принял данные. Проверь название задачи.";
  return "Сервер не смог выполнить запрос. Попробуй ещё раз.";
}

class RequestError extends Error {
  constructor(status) {
    super(statusMessage(status));
    this.status = status;
  }
}

async function request(path, method = "GET", payload, expectJSON = true) {
  const options = {
    method,
    credentials: "same-origin",
    cache: "no-store",
    headers: { Accept: "application/json" },
  };
  if (payload !== undefined) {
    options.body = JSON.stringify(payload);
    if (!path.startsWith("/auth/") && new TextEncoder().encode(options.body).length > 1024) {
      throw new Error(statusMessage(413));
    }
    options.headers["Content-Type"] = "application/json";
  }

  let response;
  try {
    response = await fetch(path, options);
  } catch {
    throw new Error("Нет связи с сервером. Проверь, что приложение запущено.");
  }
  if (!response.ok) throw new RequestError(response.status);
  if (!expectJSON || response.status === 204) return;
  try {
    return await response.json();
  } catch {
    throw new Error("Сервер вернул неожиданный ответ.");
  }
}

function syncControls() {
  refresh.disabled = state.busy;
  logout.disabled = state.busy;
  createForm.querySelector("button").disabled = state.busy || state.tasks === null;
  for (const control of taskList.querySelectorAll("button, input")) {
    control.disabled = state.busy;
  }
  byID("confirm").disabled = state.busy;
  byID("cancel").disabled = state.busy;
  editTitle.readOnly = state.busy;
}

function setView(view) {
  accountView.hidden = view === "workspace";
  workspaceView.hidden = view !== "workspace";
  authView.hidden = view !== "auth";
  sessionCheck.hidden = view !== "checking";
  document.title = view === "workspace" ? "Мои задачи · Task Manager" : "Task Manager";
}

function setPasswordVisibility(visible) {
  authPassword.type = visible ? "text" : "password";
  passwordVisibility.textContent = visible ? "Скрыть" : "Показать";
}

function syncAuthControls() {
  authSubmit.disabled = authBusy;
  authSubmit.textContent = authBusy
    ? (authMode === "login" ? "Входим…" : "Создаём аккаунт…")
    : (authMode === "login" ? "Войти" : "Создать аккаунт");
  authEmail.readOnly = authBusy;
  authPassword.readOnly = authBusy;
  passwordVisibility.disabled = authBusy;
  for (const button of authModes) button.disabled = authBusy;
}

function setAuthMode(mode) {
  authMode = mode;
  const login = mode === "login";
  byID("auth-heading").textContent = login ? "Вход в аккаунт" : "Создание аккаунта";
  byID("auth-description").textContent = login
    ? "Войди, чтобы открыть свои задачи."
    : "Укажи почту и пароль для своего списка задач.";
  authPassword.autocomplete = login ? "current-password" : "new-password";
  authPassword.value = "";
  setPasswordVisibility(false);
  showError(authError);
  showError(authNotice);
  for (const button of authModes) {
    button.setAttribute("aria-pressed", String(button.dataset.authMode === mode));
  }
  syncAuthControls();
}

function showAuth(message = "", error = "") {
  if (state.dialog !== null) closeDialog();
  state.tasks = null;
  state.filter = "all";
  newTitle.value = "";
  editTitle.value = "";
  byID("dialog-description").textContent = "";
  byID("status").textContent = "";
  authForm.reset();
  showError(errorBox);
  showError(dialogError);
  setAuthMode("login");
  render();
  setView("auth");
  showError(authNotice, message);
  showError(authError, error);
  authEmail.focus();
}

async function fetchTasks() {
  const data = await request("/tasks");
  if (data?.tasks !== null && !Array.isArray(data?.tasks)) {
    throw new Error("Сервер вернул неожиданный список задач.");
  }
  return data.tasks ?? [];
}

async function checkSession(afterLogin = false) {
  setView("checking");
  byID("session-heading").textContent = "Проверяем сессию…";
  showError(byID("session-error"));
  sessionRetry.hidden = true;
  try {
    state.tasks = await fetchTasks();
    render();
    setView("workspace");
    if (afterLogin) byID("workspace-heading").focus();
  } catch (error) {
    if (error.status === 401) {
      showAuth("", afterLogin ? "Не удалось сохранить сессию. Проверь настройки cookie и войди снова." : "");
    } else {
      byID("session-heading").textContent = "Не удалось открыть задачи";
      showError(byID("session-error"), error.message);
      sessionRetry.hidden = false;
      sessionRetry.focus();
    }
  }
}

authForm.addEventListener("submit", async (event) => {
  event.preventDefault();
  if (authBusy) return;
  const email = authEmail.value.trim();
  const password = authPassword.value.trim();
  if (!email || !password) {
    showError(authError, "Заполни почту и пароль.");
    return;
  }
  if (authMode === "register" && new TextEncoder().encode(password).length > 72) {
    showError(authError, "Пароль слишком длинный. Допустимо не более 72 байт.");
    return;
  }

  const registering = authMode === "register";
  authBusy = true;
  syncAuthControls();
  showError(authError);
  showError(authNotice);
  try {
    await request(`/auth/${authMode}`, "POST", { email, password }, false);
    if (registering) {
      setAuthMode("login");
      authEmail.value = email;
      showError(authNotice, "Аккаунт создан. Теперь войди с этим паролем.");
      authPassword.focus();
    } else {
      authPassword.value = "";
      setPasswordVisibility(false);
      await checkSession(true);
    }
  } catch (error) {
    if (!registering && error.status === 401) {
      showError(authError, "Неверная почта или пароль.");
    } else if (registering && error.status === 409) {
      showError(authError, "Аккаунт с такой почтой уже существует.");
    } else if (error.status === 400) {
      showError(authError, "Не удалось принять данные. Проверь почту и пароль.");
    } else if (error.status >= 500 && registering) {
      showError(authError, "Не удалось создать аккаунт. Если ты уже регистрировался, попробуй войти.");
    } else {
      showError(authError, error.message);
    }
  } finally {
    authBusy = false;
    syncAuthControls();
  }
});

for (const button of authModes) {
  button.addEventListener("click", () => {
    if (!authBusy && button.dataset.authMode !== authMode) setAuthMode(button.dataset.authMode);
  });
}
passwordVisibility.addEventListener("click", () => setPasswordVisibility(authPassword.type === "password"));
sessionRetry.addEventListener("click", () => checkSession());

function visibleTasks() {
  const tasks = state.tasks ?? [];
  if (state.filter === "all") return tasks;
  return tasks.filter((task) => task.done === (state.filter === "done"));
}

function render() {
  const list = visibleTasks();
  const done = state.tasks?.filter((task) => task.done).length ?? 0;
  byID("total").textContent = state.tasks?.length ?? "—";
  byID("active").textContent = state.tasks === null ? "—" : state.tasks.length - done;
  byID("done").textContent = state.tasks === null ? "—" : done;
  byID("list-heading").textContent = headings[state.filter];
  byID("list-count").textContent = state.tasks === null ? "" : `Показано: ${list.length}`;
  for (const filter of filters) {
    filter.setAttribute("aria-pressed", String(filter.dataset.filter === state.filter));
  }

  const fragment = document.createDocumentFragment();
  for (const task of list) fragment.append(renderTask(task));
  taskList.replaceChildren(fragment);
  byID("empty").hidden = list.length > 0;
  if (list.length === 0) {
    let title;
    let description;
    if (state.loading) {
      title = "Загружаем задачи…";
      description = "";
    } else if (state.tasks === null) {
      title = "Список недоступен";
      description = "Нажми «Обновить список», чтобы попробовать ещё раз.";
    } else if (state.filter === "done") {
      title = "Пока ничего не выполнено";
      description = "Здесь появятся задачи, которые ты отметишь галочкой.";
    } else if (state.filter === "active" && state.tasks.length > 0) {
      title = "Всё сделано";
      description = "Можно добавить новую задачу или посмотреть выполненные.";
    } else {
      title = "Список пока пуст";
      description = "Добавь первую задачу в поле выше.";
    }
    byID("empty-title").textContent = title;
    byID("empty-description").textContent = description;
  }
  syncControls();
}

function renderTask(task) {
  const row = byID("task-template").content.firstElementChild.cloneNode(true);
  row.dataset.id = task.id;
  row.classList.toggle("completed", task.done);
  row.querySelector(".task-title").textContent = task.title;

  const check = row.querySelector("input");
  check.checked = task.done;
  check.setAttribute("aria-label", `Выполнена: ${task.title}`);
  check.addEventListener("change", async () => {
    const done = check.checked;
    await mutate(async () => {
      await request(`/tasks/${task.id}`, "PATCH", { done });
      task.done = done;
    }, done ? "Задача выполнена" : "Задача возвращена в работу");
    focusTask(task.id, "input");
  });

  for (const mode of ["edit", "delete"]) {
    const button = row.querySelector(`.${mode}`);
    button.setAttribute("aria-label", `${button.title}: ${task.title}`);
    button.addEventListener("click", () => openDialog(mode, task));
  }
  return row;
}

function focusTask(id, selector) {
  const control = taskList.querySelector(`[data-id="${id}"] ${selector}`);
  (control ?? byID("list-heading")).focus();
}

async function mutate(action, message, box = errorBox) {
  if (state.busy) return false;
  state.busy = true;
  showError(box);
  syncControls();
  try {
    await action();
    byID("status").textContent = message;
    return true;
  } catch (error) {
    if (error.status === 401) {
      showAuth("", "Сессия закончилась. Войди снова.");
    } else {
      showError(box, error.message);
    }
    return false;
  } finally {
    state.busy = false;
    render();
  }
}

async function loadTasks() {
  if (state.busy) return;
  state.loading = true;
  render();
  await mutate(async () => {
    try {
      state.tasks = await fetchTasks();
    } finally {
      state.loading = false;
    }
  }, "Список обновлён");
}

createForm.addEventListener("submit", async (event) => {
  event.preventDefault();
  const originalTitle = newTitle.value;
  const title = originalTitle.trim();
  if (!title) {
    showError(errorBox, "Введи название задачи.");
    newTitle.focus();
    return;
  }
  const created = await mutate(async () => {
    const { id } = await request("/tasks", "POST", { title });
    if (!Number.isSafeInteger(id)) throw new Error("Сервер не вернул ID задачи. Обнови список.");
    state.tasks.push({ id, title, done: false });
    if (state.filter === "done") state.filter = "active";
  }, "Задача добавлена");
  if (created && newTitle.value === originalTitle) newTitle.value = "";
  newTitle.focus();
});

function openDialog(mode, task) {
  if (state.busy) return;
  state.dialog = { mode, task };
  const editing = mode === "edit";
  byID("dialog-title").textContent = editing ? "Переименовать задачу" : "Удалить задачу?";
  byID("dialog-description").textContent = editing ? "" : task.title;
  byID("edit-field").hidden = !editing;
  editTitle.value = editing ? task.title : "";
  editTitle.required = editing;
  const confirm = byID("confirm");
  confirm.textContent = editing ? "Сохранить" : "Удалить";
  confirm.classList.toggle("primary", editing);
  confirm.classList.toggle("danger", !editing);
  showError(dialogError);
  dialog.showModal();
  (editing ? editTitle : byID("cancel")).focus();
}

byID("dialog-form").addEventListener("submit", async (event) => {
  event.preventDefault();
  if (state.busy || state.dialog === null) return;
  const { mode, task } = state.dialog;
  const title = editTitle.value.trim();
  if (mode === "edit" && !title) {
    showError(dialogError, "Введи название задачи.");
    editTitle.focus();
    return;
  }

  const path = `/tasks/${task.id}`;
  const saved = await mutate(async () => {
    if (mode === "edit") {
      await request(path, "PATCH", { title });
      task.title = title;
    } else {
      await request(path, "DELETE");
      state.tasks = state.tasks.filter(({ id }) => id !== task.id);
    }
  }, mode === "edit" ? "Название обновлено" : "Задача удалена", dialogError);

  if (saved) closeDialog();
});

function closeDialog() {
  if (state.dialog === null) return;
  const { mode, task } = state.dialog;
  state.dialog = null;
  dialog.close();
  focusTask(task.id, `.${mode}`);
}

byID("cancel").addEventListener("click", closeDialog);
dialog.addEventListener("cancel", (event) => {
  event.preventDefault();
  if (!state.busy) closeDialog();
});

for (const filter of filters) {
  filter.addEventListener("click", () => {
    state.filter = filter.dataset.filter;
    render();
  });
}

refresh.addEventListener("click", loadTasks);
logout.addEventListener("click", async () => {
  const signedOut = await mutate(
    () => request("/auth/logout", "POST", undefined, false),
    "Выход из аккаунта",
  );
  if (signedOut) showAuth("Ты вышел из аккаунта.");
});
checkSession();
