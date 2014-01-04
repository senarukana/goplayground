/*////对n(10000000)个整数排序，采用二路归并的方法。每次总是将两个文件合并排序为一个。
string get_file_name(int count_file)
{
	stringstream s;
	s<<count_file;
	string count_file_string;
	s>>count_file_string;
	string file_name="data";
	file_name+=count_file_string;
	return file_name;
}
//用二路归并的方法将n个整数分为每个大小为per的部分。然后逐级递增的合并
void push_random_data_to_file(const string& filename ,unsigned long number)
{
		if (number<100000)
		{
			vector<int> a;
			push_rand(a,number,0,number);
			write_data_to_file(a,filename.c_str()); 
		} 
		else
		{
			vector<int> a;
			const int per=100000,n=number/per;
			push_rand(a,number%per,0,number);
			write_data_to_file(a,filename.c_str()); 
			for (int i=0;i<n;i++)
			{
				a.clear();
				push_rand(a,100000,0,100000);
				write_data_append_file(a,filename.c_str()); 
			}
		}
}
void split_data(const string& datafrom,deque<string>& file_name_array,unsigned long per,int& count_file)
{
	unsigned long position=0;
	while (true)	
	{
		vector<int> a;
		a.clear();
		//读文件中的一段数据到数组中 
		if (read_data_to_array(datafrom,a,position,per)==true)
		{
			break;
		}
		position+=per;
		//将数组中的数据在内存中排序
		sort(a.begin(),a.end());
		ofstream fout;
		string filename=get_file_name(count_file++);
		file_name_array.push_back(filename);
		fout.open(filename.c_str(),ios::in | ios::binary);
		//将排好序的数组输出到外部文件中
		write_data_to_file(a,filename.c_str());
		print_file(filename);
		fout.close();
	}
}
void sort_big_file_with_binary_merge(unsigned long n,unsigned long per)
{
	unsigned  long traverse=n/per;
	vector<int> a;
	//制造大量数据放入文件中
	cout<<"对"<<n<<"个整数进行二路归并排序，每一路的大小为"<<per<<endl
		<<"全部数据被分割放在"<<traverse<<"个文件中"<<endl;
	
	SingletonTimer::Instance();
	//将待排序文件分成小文件，在内存中排序后放到磁盘文件中
	string datafrom="data.txt";
	deque<string> file_name_array;
	int count_file=0;
	split_data(datafrom,file_name_array,per,count_file);

	SingletonTimer::Instance()->print("将待排序文件分成小文件，在内存中排序后放到磁盘文件中");
	//合并排序，二路归并的方法。
	while (file_name_array.size()>=2)
	{
		//获取两个有序文件中的内容，将其合并为一个有序的文件，直到最后合并为一个有序文件
		string file1=file_name_array.front();
		file_name_array.pop_front();
		string file2=file_name_array.front();
		file_name_array.pop_front();
		string fileout=get_file_name(count_file++);
		file_name_array.push_back(fileout);
		merge_file(file1,file2,fileout);
		print_file(fileout);
	}
	SingletonTimer::Instance()->print("获取两个有序文件中的内容，将其合并为一个有序的文件，直到最后合并为一个有序文件");
	cout<<"最终的文件中存放所有排好序的数据，其中前一百个为："<<endl;
	print_file(file_name_array.back(),100);

}*/