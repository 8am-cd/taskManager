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

async function request(path, method = "GET", payload) {
  const options = { method, headers: { Accept: "application/json" } };
  if (payload !== undefined) {
    options.body = JSON.stringify(payload);
    if (new TextEncoder().encode(options.body).length > 1024) {
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
  if (!response.ok) throw new Error(statusMessage(response.status));
  if (response.status === 204) return;
  try {
    return await response.json();
  } catch {
    throw new Error("Сервер вернул неожиданный ответ.");
  }
}

function syncControls() {
  refresh.disabled = state.busy;
  createForm.querySelector("button").disabled = state.busy || state.tasks === null;
  for (const control of taskList.querySelectorAll("button, input")) {
    control.disabled = state.busy;
  }
  byID("confirm").disabled = state.busy;
  byID("cancel").disabled = state.busy;
  editTitle.readOnly = state.busy;
}

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
    showError(box, error.message);
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
      const data = await request("/tasks");
      if (data?.tasks !== null && !Array.isArray(data?.tasks)) {
        throw new Error("Сервер вернул неожиданный список задач.");
      }
      state.tasks = data.tasks ?? [];
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
loadTasks();
