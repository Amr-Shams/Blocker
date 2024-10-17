// import axios to the file 
import axios from 'axios';

class NodeService{
    url; 
    constructor(port=5001){
        this.url = `http://localhost:${port}`;
    }
    async getPairs(){
        const response = await axios.get(`${this.url}/pairs`);
        return response.data;
    }
    async getInfo(){
        const response = await axios.get(`${this.url}/info`);
        return response.data;
    }

    async getStatus(){
        const response = await axios.get(`${this.url}/status`);
        return response.data;
    }
    async getBalance(){
        const response = await axios.get(`${this.url}/balance`);
        return response.data;
    }
}

export default NodeService;
