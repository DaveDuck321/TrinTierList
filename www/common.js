export async function PostJSON(url, Data) {
    const Request = await fetch(url, {
        method: "POST",
        headers: {
            "Content-Type": "application/json"
        },
        body: JSON.stringify(Data),
    });

    const Response = await Request.json();
    if (!Response.success)
        console.error(Response.msg);

    return Response;
}

export function PopulateCategories(categories, onchange) {
    const select = document.getElementById("category");
    for (const category of categories) {
        const option = document.createElement("option");
        option.value = category.id;
        option.innerText = category.name;

        select.appendChild(option);
    }
    
    select.addEventListener("change", onchange);
}