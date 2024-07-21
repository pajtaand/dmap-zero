// Constants
const HOST = getCurrentHost();
const POPUP_MESSAGE_TIME = 4500
const PAGE_RELOAD_INTERVAL = 2500

credentials = null

// App state management

window.onload = function() {
    loadApp().then(); 
};

async function loadApp() {
    console.log("Loading application");

    if (credentials == null) {
        showWelcomeView();
        return;
    }
    
    showUserView().then();
}

function showWelcomeView() {
    console.log("Loading welcome view");

    // load view
    document.getElementById("welcome-view").style.display = "block";
    document.getElementById("user-view").style.display = "none";
    document.getElementById("loading-view").style.display = "none";
}

async function showUserView() {
    console.log("Loading user view");

    // load user content
    await reloadAppData();

    // load view
    document.getElementById("welcome-view").style.display = "none";
    document.getElementById("user-view").style.display = "block";
    document.getElementById("loading-view").style.display = "none";

    // show tab
    showTab("agents");
}

function showTab(tabName) {
    console.log("Showing tab: " + tabName);

    // deactivate all buttons
    let agentsTabButton = document.getElementById("sidebar-tab-agents");
    let imagesTabButton = document.getElementById("sidebar-tab-images");
    let modulesTabButton = document.getElementById("sidebar-tab-modules");

    if (agentsTabButton.classList.contains('active')) {
        agentsTabButton.classList.remove("active");
    }
    if (imagesTabButton.classList.contains('active')) {
        imagesTabButton.classList.remove("active");
    }
    if (modulesTabButton.classList.contains('active')) {
        modulesTabButton.classList.remove("active");
    }

    // hide all tabs
    document.getElementById("tab-agents").style.display = "none";
    document.getElementById("tab-images").style.display = "none";
    document.getElementById("tab-modules").style.display = "none";

    // activate correct button
    document.getElementById("sidebar-tab-" + tabName).classList.add("active");

    // show correct tab
    document.getElementById("tab-" + tabName).style.display = "flex";

    // always show the top part of the page
    jumpToStart();

    // reload data
    reloadAppData().then();
}

async function reloadAppData() {
    console.log("Loading app data");

    const listAgentsRequest = fetch(`${HOST}/api/v1/agent`, {
        method: "GET",
        cache: "no-cache",
        headers: {
            "Authorization": `Basic ${credentials}`,
            "Content-Type": "application/json",
        },
    });
    const listModulesRequest = fetch(`${HOST}/api/v1/module`, {
        method: "GET",
        cache: "no-cache",
        headers: {
            "Authorization": `Basic ${credentials}`,
            "Content-Type": "application/json",
        },
    });
    const listImagesRequest = fetch(`${HOST}/api/v1/image`, {
        method: "GET",
        cache: "no-cache",
        headers: {
            "Authorization": `Basic ${credentials}`,
            "Content-Type": "application/json",
        },
    });

    const listAgentsResponse = await listAgentsRequest
    const listModulesResponse = await listModulesRequest
    const listImagesResponse = await listImagesRequest

    if (listAgentsResponse.status !== 200) {
        showError("Failed to load agents");
        return
    }
    if (listModulesResponse.status !== 200) {
        showError("Failed to load modules");
        return
    }
    if (listImagesResponse.status !== 200) {
        showError("Failed to load images");
        return
    }

    let agents = (await listAgentsResponse.json()).Agents;
    let images = (await listImagesResponse.json()).Images;
    let modules = (await listModulesResponse.json()).Modules;

    const agentListHtml = document.getElementById("list-agents");
    const imagesListHtml = document.getElementById("list-images");
    const modulesListHtml = document.getElementById("list-modules");
    const agentItemTemplate = document.getElementById("template-item-agent");
    const imageItemTemplate = document.getElementById("template-item-image");
    const moduleItemTemplate = document.getElementById("template-item-module");

    // remove all old data
    Array.from(agentListHtml.children).forEach(function(element) {
        element.remove();
    })
    Array.from(imagesListHtml.children).forEach(function(element) {
        element.remove();
    })
    Array.from(modulesListHtml.children).forEach(function(element) {
        element.remove();
    })

    // sort
    agents.sort(function (a, b) {
        return ('' + a.Name).localeCompare(b.Name);
    })
    images.sort(function (a, b) {
        return ('' + a.Name).localeCompare(b.Name);
    })
    modules.sort(function (a, b) {
        return ('' + a.Name).localeCompare(b.Name);
    })

    // get stats data
    let imagesTotalSize = 0
    let onlineAgents = 0;
    let runningModules = 0;
    let moduleInstances = 0;
    let moduleMap = {};
    modules.forEach(module => {
        moduleMap[module.ID] = 0;
        runningModules += module.IsRunning
    });
    images.forEach(image => {
        imagesTotalSize += image.Size;
    });
    agents.forEach(agent => {
        onlineAgents += agent.IsOnline && agent.IsEnrolled;
        agent.PresentModules.forEach(modID => {
            if (moduleMap[modID] !== undefined) {
                moduleMap[modID]++;
            }
            moduleInstances++;
        });
    });

    // update texts
    document.getElementById("text-agents-info").innerHTML = `Agents (${agents.length})`;
    document.getElementById("text-agents-info-2").innerHTML = `Online: ${onlineAgents}, Offline: ${agents.length-onlineAgents}`;
    document.getElementById("text-images-info").innerHTML = `Images (${images.length})`;
    document.getElementById("text-images-info-2").innerHTML = `Total size: ${getSizeFormat(imagesTotalSize)}`;
    document.getElementById("text-modules-info").innerHTML = `Modules (${modules.length})`;
    document.getElementById("text-modules-info-2").innerHTML = `Online: ${runningModules}, Offline: ${modules.length-runningModules}`;

    // insert new data
    agents.forEach(function(agent) {
        const newItemHtml = document.createElement("span");
        newItemHtml.innerHTML = agentItemTemplate.innerHTML;
        newItemHtml.style.display = "flex";
        newItemHtml.getElementsByClassName("text-horizontal")[0].innerHTML = agent.Name;
        newItemHtml.getElementsByClassName("text-horizontal-small")[0].innerHTML = agent.IsEnrolled ? (agent.IsOnline ? "Online" : "Offline") : "Not enrolled" ;
        newItemHtml.getElementsByClassName("text-horizontal-small")[0].style.color = agent.IsEnrolled ? (agent.IsOnline ? "green" : "red") : "grey";

        let revokeButton = newItemHtml.getElementsByClassName("button-revoke")[0];
        revokeButton.setAttribute("obj-id", agent.ID);
        if (!agent.IsEnrolled) {
            revokeButton.style.display = "none";
        }

        let enrollmentButton = newItemHtml.getElementsByClassName("button-enrollment")[0];
        enrollmentButton.setAttribute("obj-id", agent.ID);
        if (agent.IsEnrolled) {
            enrollmentButton.style.display = "none";
        }

        let editButton = newItemHtml.getElementsByClassName("button-edit")[0];
        editButton.setAttribute("obj-id", agent.ID);

        let deleteButton = newItemHtml.getElementsByClassName("button-delete")[0];
        deleteButton.setAttribute("obj-id", agent.ID);

        agentListHtml.appendChild(newItemHtml);
    })
    images.forEach(function(image) {
        const newItemHtml = document.createElement("span");
        newItemHtml.innerHTML = imageItemTemplate.innerHTML;
        newItemHtml.style.display = "flex";
        newItemHtml.getElementsByClassName("text-horizontal")[0].innerHTML = image.Name;
        newItemHtml.getElementsByClassName("text-horizontal-small")[0].innerHTML = `Size: ${getSizeFormat(image.Size)}`;

        let editButton = newItemHtml.getElementsByClassName("button-edit")[0];
        editButton.setAttribute("obj-id", image.ID);
        editButton.style.visibility = "hidden";

        let deleteButton = newItemHtml.getElementsByClassName("button-delete")[0];
        deleteButton.setAttribute("obj-id", image.ID);
        
        imagesListHtml.appendChild(newItemHtml);
    })
    modules.forEach(function(module) {
        const newItemHtml = document.createElement("span");
        newItemHtml.innerHTML = moduleItemTemplate.innerHTML;
        newItemHtml.style.display = "flex";
        newItemHtml.getElementsByClassName("text-horizontal")[0].innerHTML = module.Name;
        newItemHtml.getElementsByClassName("text-horizontal-small")[0].innerHTML = module.IsRunning ? `Running on ${moduleMap[module.ID]} / ${onlineAgents} agents` : `Not running`;
        newItemHtml.getElementsByClassName("text-horizontal-small")[0].style.color = module.IsRunning ? `green` : `grey`;

        let editButton = newItemHtml.getElementsByClassName("button-edit")[0];
        editButton.setAttribute("obj-id", module.ID);

        let deleteButton = newItemHtml.getElementsByClassName("button-delete")[0];
        deleteButton.setAttribute("obj-id", module.ID);

        let startButton = newItemHtml.getElementsByClassName("button-start")[0];
        startButton.setAttribute("obj-id", module.ID);
        if (module.IsRunning) {
            startButton.style.display = "none";
        }

        let stopButton = newItemHtml.getElementsByClassName("button-stop")[0];
        stopButton.setAttribute("obj-id", module.ID);
        if (!module.IsRunning) {
            stopButton.style.display = "none";
        }

        modulesListHtml.appendChild(newItemHtml);
    })
}

function onFileUpload(input) {
    if (input.files[0] == null) {
        input.parentElement.getElementsByTagName("label")[0].innerHTML = "Upload file"
    } else {
        input.parentElement.getElementsByTagName("label")[0].innerHTML = input.files[0].name;
    }
}

async function formLogin(form) {
    let username = form["input-login-username"].value;
    let password = form["input-login-password"].value;
    
    // check the credentials are correct
    fetch(`${HOST}/api/v1/agent`, {
        method: "GET",
        cache: "no-cache",
        headers: {
            "Authorization": `Basic ${btoa(username + ":" + password)}`,
            "Content-Type": "application/json",
        },
    }).then(response => {
        if (response.status === 200) {
            credentials = btoa(username + ":" + password)
            loadApp().then();
        } else if (response.status === 401) {
            showError("Wrong credentials");
        } else {
            showError("Error");
        }
    });
}

function showAddAgentForm() {
    showWindow("add-agent-popup");
    document.getElementById("form-add-agent-input-name").value = "";
    document.getElementById("form-add-agent-input-config").value = "";
}

function buttonEditAgent(button) {
    let agentID = button.getAttribute("obj-id");
    if (agentID == null) {
        showError("Invalid agent ID!");
        return;
    }

    fetch(`${HOST}/api/v1/agent/${agentID}`, {
        method: "GET",
        cache: "no-cache",
        headers: {
            "Authorization": `Basic ${credentials}`,
            "Content-Type": "application/json",
        },
    }).then(response => {
        if (response.status !== 200) {
            showError("Failed to get agent");
            return
        }

        response.json().then(agent => {
            showWindow("edit-agent-popup");
            document.getElementById("form-edit-agent").setAttribute("obj-id", agentID);
            document.getElementById("form-edit-agent-input-name").value = agent.Name;
            let cfgString = "";
            for (var key in agent.Configuration){
                cfgString += key + "=" + agent.Configuration[key] + "\n";
            } 
            document.getElementById("form-edit-agent-input-config").value = cfgString;
        });
    });
}

function buttonRevokeAgent(button) {
    let agentID = button.getAttribute("obj-id");
    if (agentID == null) {
        showError("Invalid agent ID!");
        return;
    }

    fetch(`${HOST}/api/v1/agent/${agentID}/enrollment`, {
        method: "DELETE",
        cache: "no-cache",
        headers: {
            "Authorization": `Basic ${credentials}`,
            "Content-Type": "application/json",
        },
    }).then(response => {
        if (response.status !== 200) {
            showError("Failed to remove agent enrollment");
            return
        }
        showSuccess("Agent identity successfully revoked");
        reloadAppData();
    });
}

function showEnrollmentWindow(agentID, enrollmentToken, tokenExpiresAt) {
    document.getElementById("button-enroll-agent").setAttribute("obj-id", agentID);
    document.getElementById("button-enroll-agent-delete").setAttribute("obj-id", agentID);
    if (enrollmentToken != "" && tokenExpiresAt != null) {
        if (tokenExpiresAt > new Date()) {
            document.getElementById("agent-enrollment-token-expire-time").style.color = "grey";
            document.getElementById("agent-enrollment-token-expire-time").innerHTML = `The enrollment token is valid until: ${getDateTimeFormat(new Date(tokenExpiresAt))}`;
        } else {
            document.getElementById("agent-enrollment-token-expire-time").style.color = "red";
            document.getElementById("agent-enrollment-token-expire-time").innerHTML = `The enrollment token has expired: ${getDateTimeFormat(new Date(tokenExpiresAt))}`;
        }
        document.getElementById("agent-enrollment-token").innerHTML = enrollmentToken;
        
        document.getElementById("agent-enrollment-token").style.display = "unset";
        document.getElementById("button-copy-agent-enrollment-token").style.display = "unset";
        document.getElementById("button-enroll-agent").style.display = "none";
        document.getElementById("button-enroll-agent-delete").style.display = "unset";
        document.getElementById("agent-enrollment-token-expire-time").style.display = "unset";
    } else {
        document.getElementById("agent-enrollment-token").style.display = "none";
        document.getElementById("button-copy-agent-enrollment-token").style.display = "none";
        document.getElementById("button-enroll-agent").style.display = "unset";
        document.getElementById("button-enroll-agent-delete").style.display = "none";
        document.getElementById("agent-enrollment-token-expire-time").style.display = "none";
    }
    showWindow("enroll-agent-popup");
}

async function buttonShowAgentEnrollment(button) {
    let agentID = button.getAttribute("obj-id");
    if (agentID == null) {
        showError("Invalid agent ID!");
        return;
    }

    const response = await fetch(`${HOST}/api/v1/agent/${agentID}/enrollment`, {
        method: "GET",
        cache: "no-cache",
        headers: {
            "Authorization": `Basic ${credentials}`,
            "Content-Type": "application/json",
        },
    });
    let enrollmentToken = "";
    let tokenExpiresAt = "";

    if (response.status !== 200) {
        if (response.status !== 404) {
            showError("Failed to get agent enrollment");
            return
        }
    } else {
        let data = await response.json();
        enrollmentToken = data.JWT;
        tokenExpiresAt = new Date(data.ExpiresAt);
    }

    showEnrollmentWindow(agentID, enrollmentToken, tokenExpiresAt);
}

function buttonEnrollAgent(button) {
    let agentID = button.getAttribute("obj-id");
    if (agentID == null) {
        showError("Invalid agent ID!");
        return;
    }

    fetch(`${HOST}/api/v1/agent/${agentID}/enrollment`, {
        method: "POST",
        cache: "no-cache",
        headers: {
            "Authorization": `Basic ${credentials}`,
            "Content-Type": "application/json",
        },
    }).then(response => {
        if (response.status !== 200) {
            showError("Failed to enroll agent");
            return
        }

        response.json().then(data => {
            let enrollmentToken = data.JWT;
            let tokenExpiresAt = new Date(data.ExpiresAt)
            showEnrollmentWindow(agentID, enrollmentToken, tokenExpiresAt);
        });
    });
}

function buttonEnrollAgentDelete(button) {
    let agentID = button.getAttribute("obj-id");
    if (agentID == null) {
        showError("Invalid agent ID!");
        return;
    }

    fetch(`${HOST}/api/v1/agent/${agentID}/enrollment`, {
        method: "DELETE",
        cache: "no-cache",
        headers: {
            "Authorization": `Basic ${credentials}`,
            "Content-Type": "application/json",
        },
    }).then(response => {
        if (response.status !== 200) {
            showError("Failed to remove agent enrollment");
            return
        }
        showEnrollmentWindow(agentID, "", null);
    });
}

function showUploadImageForm() {
    showWindow("add-image-popup");
    document.getElementById("add-image-popup").getElementsByTagName("label")[0].innerHTML = "Upload file";
    document.getElementById("form-add-image-input-name").value = "";
    document.getElementById("form-add-image-input-file").value = null;
}

async function updateImageSelection() {
    const response = await fetch(`${HOST}/api/v1/image`, {
        method: "GET",
        cache: "no-cache",
        headers: {
            "Authorization": `Basic ${credentials}`,
            "Content-Type": "application/json",
        },
    });

    if (response.status !== 200) {
        showError("Failed to get images");
        return
    }
    const images = (await response.json()).Images;
    
    // remove old items
    const selectListHtml1 = document.getElementById("form-add-module-input-image");
    const selectListHtml2 = document.getElementById("form-edit-module-input-image");
    Array.from(selectListHtml1.children).forEach(function(element) {
        element.remove();
    })
    Array.from(selectListHtml2.children).forEach(function(element) {
        element.remove();
    })

    // add default
    const emptyItemHtml = document.createElement("option");
    emptyItemHtml.setAttribute("value", "");
    emptyItemHtml.setAttribute("disabled", "");
    emptyItemHtml.setAttribute("selected", "");
    emptyItemHtml.innerHTML = "Select image";
    selectListHtml1.appendChild(emptyItemHtml);
    selectListHtml2.appendChild(emptyItemHtml.cloneNode(true));

    // add images
    for (let i = 0; i < images.length; i++) {
        const imageItemHtml = document.createElement("option");
        imageItemHtml.setAttribute("value", images[i].ID);
        imageItemHtml.innerHTML = images[i].Name;
        selectListHtml1.appendChild(imageItemHtml);
        selectListHtml2.appendChild(imageItemHtml.cloneNode(true));
    }
}

function showSetupModuleForm() {
    document.getElementById("form-add-module-input-name").value = "";
    document.getElementById("form-add-module-input-image").value = "";
    document.getElementById("form-add-module-input-config").value = "";
    updateImageSelection().then(showWindow("add-module-popup"));
}

function formAddAgent(form) {
    let nameInput = form["form-add-agent-input-name"];
    let configInput = form["form-add-agent-input-config"];
    
    // validation
    let groups = /^((\w+)=(.+))?(\n(\w+)=(.+))*(\n)?$/.exec(configInput.value);
    if (groups == null) {
        showError("Configuration is in invalid format!");
        return;
    }

    let configuration = {};
    for (let i = 2; i < groups.length; i += 3) {
        configuration[groups[i]] = groups[i+1];
    }

    fetch(`${HOST}/api/v1/agent`, {
        method: "POST",
        cache: "no-cache",
        headers: {
            "Authorization": `Basic ${credentials}`,
            "Content-Type": "application/json",
        },
        body: JSON.stringify({
            Name: nameInput.value,
            Configuration: configuration,
        }),
    }).then(response => {
        if (response.status !== 200) {
            showError("Failed to create agent");
            return
        }
        showSuccess("Agent successfully created");
        reloadAppData();

        // hide window
        hideWindow("add-agent-popup");
    });
}

function formEditAgent(form) {
    let agentID = form.getAttribute("obj-id");
    if (agentID == null) {
        showError("Invalid agent ID!");
        return;
    }

    let nameInput = form["form-edit-agent-input-name"];
    let configInput = form["form-edit-agent-input-config"];
    
    // validation
    let groups = /^((\w+)=(.+))?(\n(\w+)=(.+))*(\n)?$/.exec(configInput.value);
    if (groups == null) {
        showError("Configuration is in invalid format!");
        return;
    }

    let configuration = {};
    for (let i = 2; i < groups.length; i += 3) {
        configuration[groups[i]] = groups[i+1];
    }

    fetch(`${HOST}/api/v1/agent/${agentID}`, {
        method: "PATCH",
        cache: "no-cache",
        headers: {
            "Authorization": `Basic ${credentials}`,
            "Content-Type": "application/json",
        },
        body: JSON.stringify({
            Name: nameInput.value,
            Configuration: configuration,
        }),
    }).then(response => {
        if (response.status !== 200) {
            showError("Failed to update agent");
            return
        }
        showSuccess("Agent successfully updated");
        reloadAppData();

        // hide window
        hideWindow("edit-agent-popup");
    });
}

function formSetupModule(form) {
    let nameInput = form["form-add-module-input-name"];
    let imageInput = form["form-add-module-input-image"];
    let configInput = form["form-add-module-input-config"];
    
    // validation
    if (imageInput.value === "") {
        showError("Image has to be selected!");
        return;
    }

    let groups = /^((\w+)=(.+))?(\n(\w+)=(.+))*(\n)?$/.exec(configInput.value);
    if (groups == null) {
        showError("Configuration is in invalid format!");
        return;
    }

    let configuration = {};
    for (let i = 2; i < groups.length; i += 3) {
        configuration[groups[i]] = groups[i+1];
    }

    fetch(`${HOST}/api/v1/module`, {
        method: "POST",
        cache: "no-cache",
        headers: {
            "Authorization": `Basic ${credentials}`,
            "Content-Type": "application/json",
        },
        body: JSON.stringify({
            Name: nameInput.value,
            Image: imageInput.value,
            Configuration: configuration,
        }),
    }).then(response => {
        if (response.status !== 200) {
            showError("Failed to create module");
            return
        }
        showSuccess("Module successfully set up");
        reloadAppData();

        // hide window
        hideWindow("add-module-popup");
    });
}

async function buttonEditModule(button) {
    let moduleID = button.getAttribute("obj-id");
    if (moduleID == null) {
        showError("Invalid module ID!");
        return;
    }

    fetch(`${HOST}/api/v1/module/${moduleID}`, {
        method: "GET",
        cache: "no-cache",
        headers: {
            "Authorization": `Basic ${credentials}`,
            "Content-Type": "application/json",
        },
    }).then(response => {
        if (response.status !== 200) {
            showError("Failed to get module");
            return
        }

        response.json().then(module => {
            updateImageSelection().then(
                function() {
                    document.getElementById("form-edit-module").setAttribute("obj-id", moduleID);
                    document.getElementById("form-edit-module-input-name").value = module.Name;
                    document.getElementById("form-edit-module-input-image").value = module.Image;
                    let cfgString = "";
                    for (var key in module.Configuration){
                        cfgString += key + "=" + module.Configuration[key] + "\n";
                    } 
                    document.getElementById("form-edit-module-input-config").value = cfgString;
                    showWindow("edit-module-popup");
                }
            );
        });
    });
}

async function buttonStartModule(button) {
    let moduleID = button.getAttribute("obj-id");
    if (moduleID == null) {
        showError("Invalid module ID!");
        return;
    }

    fetch(`${HOST}/api/v1/module/${moduleID}/start`, {
        method: "POST",
        cache: "no-cache",
        headers: {
            "Authorization": `Basic ${credentials}`,
            "Content-Type": "application/json",
        },
    }).then(response => {
        if (response.status !== 200) {
            showError("Failed to start module");
            return
        }
        showSuccess("Module successfully started");
        reloadAppData();
    });
}

async function buttonStopModule(button) {
    let moduleID = button.getAttribute("obj-id");
    if (moduleID == null) {
        showError("Invalid module ID!");
        return;
    }

    fetch(`${HOST}/api/v1/module/${moduleID}/stop`, {
        method: "POST",
        cache: "no-cache",
        headers: {
            "Authorization": `Basic ${credentials}`,
            "Content-Type": "application/json",
        },
    }).then(response => {
        if (response.status !== 200) {
            showError("Failed to stop module");
            return
        }
        showSuccess("Module successfully stopped");
        reloadAppData();
    });
}

function formEditModule(form) {
    let nameInput = form["form-edit-module-input-name"];
    let imageInput = form["form-edit-module-input-image"];
    let configInput = form["form-edit-module-input-config"];

    let moduleID = form.getAttribute("obj-id");
    if (moduleID == null) {
        showError("Invalid module ID!");
        return;
    }
    
    // validation
    if (imageInput.value === "") {
        showError("Image has to be selected!");
        return;
    }

    let groups = /^((\w+)=(.+))?(\n(\w+)=(.+))*(\n)?$/.exec(configInput.value);
    if (groups == null) {
        showError("Configuration is in invalid format!");
        return;
    }

    let configuration = {};
    for (let i = 2; i < groups.length; i += 3) {
        configuration[groups[i]] = groups[i+1];
    }

    fetch(`${HOST}/api/v1/module/${moduleID}`, {
        method: "PATCH",
        cache: "no-cache",
        headers: {
            "Authorization": `Basic ${credentials}`,
            "Content-Type": "application/json",
        },
        body: JSON.stringify({
            Name: nameInput.value,
            Image: imageInput.value,
            Configuration: configuration,
        }),
    }).then(response => {
        if (response.status !== 200) {
            showError("Failed to update module");
            return
        }
        showSuccess("Module successfully updated");
        reloadAppData();

        // hide window
        hideWindow("edit-module-popup");
    });
}

function formUploadImage(form) {
    let nameInput = form["form-add-image-input-name"];
    let fileInput = form["form-add-image-input-file"];
    let file = fileInput.files[0];

    // validation
    if (file == null) {
        showError("No image file provided!");
        return;
    }

    // hide window
    hideWindow("add-image-popup");

    // prepare upload template
    const imageItemTemplate = document.getElementById("template-item-image");
    const imageListHtml = document.getElementById("list-images");
    const newItemHtml = document.createElement("span");
    newItemHtml.innerHTML = imageItemTemplate.innerHTML;
    newItemHtml.style.display = "flex";
    newItemHtml.getElementsByClassName("text-horizontal")[0].innerHTML = nameInput.value;
    imageListHtml.appendChild(newItemHtml);

    const xhr = new XMLHttpRequest();
    xhr.open("POST", `${HOST}/api/v1/image`, true);
    xhr.setRequestHeader("Authorization", `Basic ${credentials}`);

    xhr.upload.addEventListener('progress', (event) => {
        if (event.lengthComputable) {
            const percentComplete = (event.loaded / event.total) * 100;
            if (event.loaded ==  event.total) {
                newItemHtml.getElementsByClassName("text-horizontal-small")[0].innerHTML = "Waiting for server to process the file";
            } else {
                newItemHtml.getElementsByClassName("text-horizontal-small")[0].innerHTML = `Uploaded ${getSizeFormat(event.loaded)}/${getSizeFormat(event.total)} (${percentComplete.toFixed(2)}%)`;
            }
        }
    });

    xhr.onload = () => {
        if (xhr.status === 200) {
            showSuccess("Image uploaded successfully")
        } else {
            showError("File upload failed")
        }

        reloadAppData().then();
    };

    xhr.onerror = () => {
        imageListHtml.removeChild(imageListHtml.lastElementChild);
        showError("Error during the upload")
    };

    const formData = new FormData();
    formData.append('file', file);
    formData.append('name', nameInput.value);
    xhr.send(formData);
}

function buttonDeleteAgent(form) {
    let agentID = form.getAttribute("obj-id");
    if (agentID == null) {
        showError("Invalid agent ID!");
        return;
    }
    fetch(`${HOST}/api/v1/agent/${agentID}`, {
        method: "DELETE",
        cache: "no-cache",
        headers: {
            "Authorization": `Basic ${credentials}`,
            "Content-Type": "application/json",
        },
    }).then(response => {
        if (response.status !== 200) {
            showError("Failed to delete agent");
            return
        }
        showSuccess("Agent successfully deleted");
        reloadAppData(); 
    });
}

function buttonDeleteImage(form) {
    let imageID = form.getAttribute("obj-id");
    if (imageID == null) {
        showError("Invalid image ID!");
        return;
    }
    fetch(`${HOST}/api/v1/image/${imageID}`, {
        method: "DELETE",
        cache: "no-cache",
        headers: {
            "Authorization": `Basic ${credentials}`,
            "Content-Type": "application/json",
        },
    }).then(response => {
        if (response.status !== 200) {
            showError("Failed to delete image");
            return
        }
        showSuccess("Image successfully deleted");
        reloadAppData(); 
    });
}

function buttonDeleteModule(form) {
    let moduleID = form.getAttribute("obj-id");
    if (moduleID == null) {
        showError("Invalid module ID!");
        return;
    }
    fetch(`${HOST}/api/v1/module/${moduleID}`, {
        method: "DELETE",
        cache: "no-cache",
        headers: {
            "Authorization": `Basic ${credentials}`,
            "Content-Type": "application/json",
        },
    }).then(response => {
        if (response.status !== 200) {
            showError("Failed to delete module");
            return
        }
        showSuccess("Module successfully deleted");
        reloadAppData(); 
    });
}

function checkValidImage(htmlSelect) {
    if (htmlSelect.value === "") {
        htmlSelect.setCustomValidity("Image must be selected!");
    } else {
        htmlSelect.setCustomValidity("");
    }
}

function copyInnerHTML(elementID) {
    const text = document.getElementById(elementID).innerHTML;
    navigator.clipboard.writeText(text);
    showInfo("Text copied to clipboard");
}

function getCurrentHost() {
    var currentURL = window.location.href;
    if (currentURL.endsWith('/') && currentURL !== '/') {
        currentURL = currentURL.slice(0, -1);
    }
    return currentURL;
}

// INTERACTIVE CSS ELEMENTS

let lastYPos = window.scrollY;
window.onscroll = function() {
    var width = (window.innerWidth > 0) ? window.innerWidth : screen.width;

    // show sidebar
    let currentYPos = window.scrollY;
    let sidebarHtml = document.getElementById("sidebar");
    if (width >= 550 && sidebarHtml != null) {
        if (lastYPos > currentYPos) {
            sidebarHtml.style.top = "0px";
        } else {
            sidebarHtml.style.top = "-60px";
        }
    }
    lastYPos = currentYPos;

    // show jump button
    let jumpToStartHtml = document.getElementById("jump-to-start");
    if (jumpToStartHtml != null) {
        if (document.body.scrollTop > 20 || document.documentElement.scrollTop > 20) {
            jumpToStartHtml.style.bottom = "20px";
        } else {
            jumpToStartHtml.style.bottom = "-60px";
        }
    }
}

function jumpToStart() {
    document.body.scrollTop = 0;
    document.documentElement.scrollTop = 0;
}

function showMessage(message, type) {
    console.log(`Showing popup message, message: ${message}, type: ${type}`);

    let messageHtml = document.getElementById("pop-message");
    messageHtml.innerHTML = message;
    messageHtml.classList.remove(...messageHtml.classList);
    messageHtml.classList.add(type)
    if (!messageHtml.classList.contains('show')) {
        messageHtml.classList.add("show")
        setTimeout(function() {
            messageHtml.innerHTML = "";
            messageHtml.className = messageHtml.className.replace("show", "");
        }, POPUP_MESSAGE_TIME);
    }
}

function showInfo(message) {
    showMessage(message, "info");
}

function showSuccess(message) {
    showMessage(message, "success");
}

function showError(message) {
    showMessage(message, "error");
}

function showWindow(htmlElementID) {
    document.getElementById(htmlElementID).style.display = "block";
}

function hideWindow(htmlElementID) {
    document.getElementById(htmlElementID).style.display = "none";
}

window.onkeydown = function (event) {
    if (event.key === "Escape") {
        ["add-agent-popup", "add-image-popup", "add-module-popup", "edit-module-popup", "edit-agent-popup", "edit-module-popup", "enroll-agent-popup"].forEach(function(htmlElementID) {
            hideWindow(htmlElementID);
        });
    }
}

window.onclick = function (event) {
    ["add-agent-popup", "add-image-popup", "add-module-popup", "edit-module-popup", "edit-agent-popup", "edit-module-popup", "enroll-agent-popup"].forEach(function(htmlElementID) {
        if (event.target === document.getElementById(htmlElementID)) {
            hideWindow(htmlElementID);
        }
    });
}

// other helper function

function getDateTimeFormat(date) {
    return date.getDate() + "." + (date.getMonth() + 1) + "." + date.getFullYear() + (date.getHours() < 10 ? " 0" : " ") + date.getHours() + (date.getMinutes() < 10 ? ":0" : ":") + date.getMinutes();
}

function getSizeFormat(bytes, decimals = 2) {
    // source: https://stackoverflow.com/questions/15900485/correct-way-to-convert-size-in-bytes-to-kb-mb-gb-in-javascript
    if (!+bytes) return '0 Bytes'

    const k = 1024
    const dm = decimals < 0 ? 0 : decimals
    const sizes = ['Bytes', 'KiB', 'MiB', 'GiB', 'TiB', 'PiB', 'EiB', 'ZiB', 'YiB']

    const i = Math.floor(Math.log(bytes) / Math.log(k))

    return `${parseFloat((bytes / Math.pow(k, i)).toFixed(dm))} ${sizes[i]}`
}
    