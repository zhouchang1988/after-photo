export namespace main {
	
	export class DirStats {
	    jpg: number;
	    raw: number;
	    video: number;
	    other: number;
	    totalSize: number;
	
	    static createFrom(source: any = {}) {
	        return new DirStats(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.jpg = source["jpg"];
	        this.raw = source["raw"];
	        this.video = source["video"];
	        this.other = source["other"];
	        this.totalSize = source["totalSize"];
	    }
	}

}

