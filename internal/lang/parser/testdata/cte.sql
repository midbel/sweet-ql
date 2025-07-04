with persons1 as (select firstname, lastname from employees) select firstname from persons1;

with persons2 as (
	select id, firstname, lastname from employees
), emails1 as (
	select id, mail from emails
)
select t1.firstname, t2.mail from persons2 t1 join emails1 t2 on t1.id=t2.id;

with persons3 as (
	select id, firstname, lastname from employees
), emails2 as (
	select id, mail from emails
), contacts as (
	select 
		t1.firstname, 
		t2.mail 
	from persons3 t1 
	join emails2 t2 on t1.id=t2.id
	join domains d on t2.scheme=d.scheme
)
select * from contacts;